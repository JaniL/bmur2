const moment = require("moment");

const nowPlaying = (bot, lastfm) => msg => {
  const chatId = msg.chat.id;
  const {
    track: [nowPlaying]
  } = lastfm.getTracks();

  const formatTrack = ({ artist, name, url, date }) => {
    const formatDate = date => moment.unix(Number(date.uts)).fromNow();
    return `[${artist["#text"]} - ${name}](${url})${
      date ? ` (${formatDate(date)})` : ""
    }`;
  };
  bot.sendMessage(chatId, formatTrack(nowPlaying), {
    parse_mode: "Markdown",
    disable_web_page_preview: true
  });
};

module.exports = nowPlaying;
