function sizeConvert(size, i = 0) {
  if (size < 1024) {
    let postfix = "";
    switch (i) {
      case 0:
        postfix = "B";
        break;
      case 1:
        postfix = "KB";
        break;
      case 2:
        postfix = "MB";
        break;
      case 3:
        postfix = "GB";
        break;
    }
    return Math.round(size * 100) / 100 + " " + postfix;
  }
  return sizeConvert(size / 1024, i + 1);
}

function speedConvert(speed, i = 0) {
  if (speed < 1000) {
    let postfix = "";
    switch (i) {
      case 0:
        postfix = "B";
        break;
      case 1:
        postfix = "KB";
        break;
      case 2:
        postfix = "MB";
        break;
      case 3:
        postfix = "GB";
        break;
    }
    return Math.round(speed * 10) / 10 + " " + postfix + "/s";
  }
  return speedConvert(speed / 1000, i + 1);
}

function etaConvert(eta) {
  if (eta >= 8640000) {
    return "∞";
  }
  const second = eta % 60;
  const minute = Math.floor(eta / 60) % 60;
  const hour = Math.floor(eta / 3600) % 24;
  const day = Math.floor(eta / 84000) % 30;

  const secondTag = second > 0 ? second + "s" : "";
  const minuteTag = minute > 0 ? minute + "m" : "";
  const hourTag = hour > 0 ? hour + "h" : "";
  const dayTag = day > 0 ? day + "d" : "";

  const arr = [dayTag, hourTag, minuteTag, secondTag];
  let i = 0;
  return arr.reduce((acc, v) => {
    if (v != "" && i < 2) {
      acc += " " + v;
      i++;
    }
    return acc;
  }, "");
}

export default {
  sizeConvert,
  speedConvert,
  etaConvert,
};
