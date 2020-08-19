import axios from "axios";

function AddMagnetLink(link) {
  let data = new FormData();
  data.append("urls", link);
  return axios.post("/magnet", data);
}

function GetInitData() {
  return axios.get("/torrents");
}

function StopTorrent(hashes) {
  axios.get("/torrents/pause", {
    params: {
      hashes: hashes,
    },
  });
}

function ResumeTorrent(hashes) {
  axios.get("/torrents/resume", {
    params: {
      hashes: hashes,
    },
  });
}

function DeleteTorrent(hashes) {
  axios.delete("torrents/delete", {
    params: {
      hashes: hashes,
    },
  });
}

function EncodeVideo(video) {
  axios.post("http://localhost:8081/encode/" + video);
}

function EncodeProgress(items, cb) {
  setInterval(() => {
    let param = "";
    this.items.forEach((e) => {
      param = "movies=" + e.name + "&";
    });
    axios
      .get("http://localhost:8082/progress?" + param)
      .then((r) => cb(r.data));
  }, 1500);
}

let socket;
function InitWebSocket(cb) {
  socket = new WebSocket("ws://localhost:8081/torrents/ws");

  // Listen for messages
  socket.onmessage = function (event) {
    console.log("Message from server ", event.data);
    const msg = JSON.parse(event.data);
    cb(msg);
  };
}

function CloseWB() {
  socket.close();
}

export default {
  GetInitData,
  StopTorrent,
  ResumeTorrent,
  DeleteTorrent,
  InitWebSocket,
  EncodeVideo,
  EncodeProgress,
  AddMagnetLink,
  CloseWB,
};
