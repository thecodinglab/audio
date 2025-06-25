/** @type {AudioContext | null} */
let context = null;
/** @type {RTCPeerConnection | null} */
let conn = null;

/** @type {HTMLCanvasElement} */
const canvas = document.querySelector("#canvas");
const canvasCtx = canvas.getContext("2d");

function ensureAudioContext() {
  if (!context) {
    context = new AudioContext();
  }
  if (context.state === 'suspended') {
    context.resume();
  }
}

/** @param {AudioNode} destination */
function connect(destination) {
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
    const source = context.createMediaStreamSource(event.streams[0]);
    source.connect(destination);
  });

  conn.addTransceiver("audio", {
    direction: "recvonly",
  });
}

function play() {
  ensureAudioContext();

  const analyzer = context.createAnalyser();
  analyzer.fftSize = 512;
  analyzer.minDecibels = -90;
  analyzer.maxDecibels = -10;
  analyzer.smoothingTimeConstant = 0.85;
  // analyzer.connect(context.destination);

  connect(analyzer);

  const bufferLength = analyzer.frequencyBinCount;
  const visualizeLength = bufferLength - 192; // TODO
  const data = new Uint8Array(bufferLength);

  const width = canvas.width;
  const height = canvas.height;
  canvasCtx.clearRect(0, 0, width, height);

  const draw = () => {
    requestAnimationFrame(draw);

    analyzer.getByteFrequencyData(data);
    // console.log(data);

    canvasCtx.fillStyle = "rgb(0, 0, 0)";
    canvasCtx.fillRect(0, 0, width, height);

    const barWidth = width / visualizeLength - 4;
    let x = 0;

    for (let i = 0; i < visualizeLength; i++) {
      const val = data[i] / 256.0 + 0.05;
      const barHeight = 0.5 * height * val;

      canvasCtx.fillStyle = `rgb(${127 + val * 128},0,0)`;

      canvasCtx.beginPath();
      canvasCtx.roundRect(x + 2, 0.5 * height - barHeight, barWidth, 2 * barHeight, barWidth / 2);
      canvasCtx.fill();

      x += barWidth + 4;
    }
  };
  draw();
}

document.querySelector("#play").addEventListener("click", play);
