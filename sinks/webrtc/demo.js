/** @type {AudioContext | null} */
let context = null;
/** @type {RTCPeerConnection | null} */
let conn = null;

/** @type {HTMLCanvasElement} */
const canvas = document.querySelector("#canvas");

const resize = () => {
  const size = Math.min(document.body.clientWidth, document.body.clientHeight, 300)
  canvas.width = size;
  canvas.height = size;
};

resize();
window.addEventListener("resize", () => resize());

const canvasCtx = canvas.getContext("2d");

/** @type {HTMLAudioElement} */
const audio = document.querySelector("#audio");
audio.volume = 0.005;

function ensureAudioContext() {
  if (!context) {
    context = new AudioContext();
  }
  if (context.state === 'suspended') {
    context.resume().catch(console.error);
  }
}

function connect() {
  if (conn) {
    conn.close();
  }

  conn = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }]
  });

  conn.addEventListener("negotiationneeded", () => {
    conn.createOffer().then((offer) => conn.setLocalDescription(offer));
  });

  conn.addEventListener("icecandidate", (e) => {
    if (!e.candidate) {
      fetch("/webrtc", { method: "POST", body: JSON.stringify(conn.localDescription) })
        .then((res) => res.json())
        .then((answer) => conn.setRemoteDescription(answer))
        .catch(() => setTimeout(connect, 500));
    }
  });

  conn.addEventListener("connectionstatechange", () => {
    console.log("[WebRTC] connection state:", conn.connectionState);
    if (conn.connectionState === "disconnected" || conn.connectionState === "failed") {
      connect();
    }
  });

  conn.addEventListener("track", (event) => {
    audio.srcObject = event.streams[0];
    audio.play();
  });

  conn.addTransceiver("audio", {
    direction: "recvonly",
  });
}

/** 
 * @param {number} val
 * @param {number} idx
 * @param {number} total
 */
function drawSection(val, idx, total) {
  const radius = 0.375 * canvas.height;
  const angle = Math.PI * (idx / total);
  const hue = 360.0 * (idx / total);

  const height = 2 * Math.PI * radius / total;
  const width = radius * val;
  const x = radius - width / 2;
  const y = -0.5 * height;

  canvasCtx.save()
  canvasCtx.fillStyle = `hsl(${hue}, ${30 + val * 70}%, 50%)`;
  canvasCtx.translate(canvas.width / 2, canvas.height / 2);
  canvasCtx.rotate(-0.5 * Math.PI);

  canvasCtx.rotate(angle);
  canvasCtx.beginPath();
  canvasCtx.roundRect(x, y, width, height, width / 2);
  canvasCtx.fill();

  canvasCtx.rotate(-2 * angle);
  canvasCtx.beginPath();
  canvasCtx.roundRect(x, y, width, height, width / 2);
  canvasCtx.fill();

  canvasCtx.restore();
}

function play() {
  ensureAudioContext();

  const gain = context.createGain();
  gain.gain.setValueAtTime(0.01, context.currentTime);
  gain.connect(context.destination);

  const analyzer = context.createAnalyser();
  analyzer.fftSize = 1 << 12;
  analyzer.minDecibels = -90;
  analyzer.maxDecibels = -10;
  analyzer.smoothingTimeConstant = 0.85;
  analyzer.connect(gain);

  const source = context.createMediaElementSource(audio);
  source.connect(analyzer);

  connect();

  const bufferLength = analyzer.frequencyBinCount;
  const visualizeLength = bufferLength / 1.5;
  const data = new Uint8Array(bufferLength);

  const draw = () => {
    requestAnimationFrame(draw);

    analyzer.getByteFrequencyData(data);

    const stride = 1 << 5, power = 1;
    canvasCtx.clearRect(0, 0, canvas.width, canvas.height);

    for (let i = 0; i < visualizeLength; i += stride) {
      let val = 0;
      for (let j = 0; j < stride; j++) {
        val += Math.pow(data[i + j] / 256.0, power);
      }
      val = val / stride + 0.01;

      drawSection(val, i / stride, visualizeLength / stride);
    }
  };
  draw();
}

play();
