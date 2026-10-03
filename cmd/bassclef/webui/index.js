const recordBtn = document.getElementById('recordBtn');
const statusEl = document.getElementById('status');
const resultsEl = document.getElementById('results');

const RECORD_DURATION_MS = 10000;
const RECORD_SAMPLE_RATE = 44100; // must match the catalog's ingest rate

function renderError(message) {
  statusEl.innerHTML = '<div role="alert" data-variant="error"><strong>Error.</strong> ' + message + '</div>';
}

function renderResults(matches) {
  resultsEl.innerHTML = '';

  if (matches.length === 0) {
    statusEl.innerHTML = '<div role="alert"><strong>No match found.</strong></div>';
    return;
  }

  statusEl.innerHTML = '';

  matches.forEach((m, i) => {
    const card = document.createElement('article');
    card.className = 'card match-card';

    const coverHTML = m.cover_url
      ? '<img class="match-cover" src="' + m.cover_url + '" alt="">'
      : '<div class="match-cover"></div>';

    let infoHTML = '<div class="match-info">';
    infoHTML += '<h3>' + (i + 1) + '. ' + m.title + ' - ' + m.artist + '</h3>';
    infoHTML += '<p>score=' + m.score + ', confidence=' + m.confidence.toFixed(4) + '</p>';
    if (m.album) infoHTML += '<p>' + m.album + '</p>';
    if (m.spotify_url) infoHTML += '<p><a href="' + m.spotify_url + '" target="_blank">Open in Spotify</a></p>';
    infoHTML += '</div>';

    card.innerHTML = coverHTML + infoHTML;
    resultsEl.appendChild(card);
  });
}

function setBusy(busy) {
  recordBtn.disabled = busy;
  if (busy) recordBtn.setAttribute('aria-busy', 'true');
  else recordBtn.removeAttribute('aria-busy');
}

// recognize uploads body (a Blob) as the raw request body and renders the
// response.
async function recognize(body, contentType) {
  statusEl.innerHTML = '<div aria-busy="true">Recognizing...</div>';
  resultsEl.innerHTML = '';
  setBusy(true);

  try {
    const resp = await fetch('/recognize', {
      method: 'POST',
      headers: { 'Content-Type': contentType },
      body,
    });

    const data = await resp.json();

    if (!resp.ok) {
      renderError(data.error || ('request failed with status ' + resp.status));
      return;
    }

    renderResults(data.matches);
  } catch (err) {
    renderError(err.message);
  } finally {
    setBusy(false);
  }
}

// encodeWAV packs 16-bit PCM mono samples into a WAV Blob. Browsers'
// MediaRecorder only produces webm/opus, which the backend doesn't
// accept, so recording goes through the Web Audio API instead and is
// encoded to WAV by hand here.
function encodeWAV(samples, sampleRate) {
  const buffer = new ArrayBuffer(44 + samples.length * 2);
  const view = new DataView(buffer);

  function writeString(offset, s) {
    for (let i = 0; i < s.length; i++) view.setUint8(offset + i, s.charCodeAt(i));
  }

  writeString(0, 'RIFF');
  view.setUint32(4, 36 + samples.length * 2, true);
  writeString(8, 'WAVE');
  writeString(12, 'fmt ');
  view.setUint32(16, 16, true);
  view.setUint16(20, 1, true); // PCM
  view.setUint16(22, 1, true); // mono
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, sampleRate * 2, true); // byte rate
  view.setUint16(32, 2, true); // block align
  view.setUint16(34, 16, true); // bits per sample
  writeString(36, 'data');
  view.setUint32(40, samples.length * 2, true);

  let offset = 44;
  for (let i = 0; i < samples.length; i++, offset += 2) {
    const s = Math.max(-1, Math.min(1, samples[i]));
    view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7fff, true);
  }

  return new Blob([view], { type: 'audio/wav' });
}

async function recordAudio(durationMs) {
  const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
  const audioCtx = new (window.AudioContext || window.webkitAudioContext)({ sampleRate: RECORD_SAMPLE_RATE });
  const source = audioCtx.createMediaStreamSource(stream);
  const processor = audioCtx.createScriptProcessor(4096, 1, 1);

  const chunks = [];

  source.connect(processor);
  processor.connect(audioCtx.destination);

  processor.onaudioprocess = (e) => {
    chunks.push(new Float32Array(e.inputBuffer.getChannelData(0)));
  };

  const sampleRate = audioCtx.sampleRate;

  await new Promise((resolve) => setTimeout(resolve, durationMs));

  processor.disconnect();
  source.disconnect();
  stream.getTracks().forEach((t) => t.stop());
  await audioCtx.close();

  let totalLength = 0;
  for (const c of chunks) totalLength += c.length;

  const samples = new Float32Array(totalLength);
  let offset = 0;
  for (const c of chunks) {
    samples.set(c, offset);
    offset += c.length;
  }

  return encodeWAV(samples, sampleRate);
}

recordBtn.addEventListener('click', async () => {
  setBusy(true);
  statusEl.innerHTML = '<div aria-busy="true">Recording for 10 seconds...</div>';
  resultsEl.innerHTML = '';

  try {
    const wavBlob = await recordAudio(RECORD_DURATION_MS);
    await recognize(wavBlob, 'audio/wav');
  } catch (err) {
    setBusy(false);
    renderError(err.message);
  }
});
