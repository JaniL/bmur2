const moment = require("moment");

const recentTracks = (bot, lastfm) => msg => {
  const chatId = msg.chat.id;

  const { track: tracks } = lastfm.getTracks();

  const formatTrack = ({ artist, name, url, date }) => {
    const formatDate = date => moment.unix(Number(date.uts)).fromNow();
    return `- [${artist["#text"]} - ${name}](${url})${
      date ? ` (${formatDate(date)})` : ""
    }`;
  };
  const formatTracks = tracks => tracks.map(formatTrack).join("\n");

  bot.sendMessage(chatId, formatTracks(tracks), {
    parse_mode: "Markdown",
    disable_web_page_preview: true
  });
};

module.exports = recentTracks;
