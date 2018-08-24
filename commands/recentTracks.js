const recentTracks = (bot, lastfm) => msg => {
  const chatId = msg.chat.id;

  const { track: tracks } = lastfm.getTracks();

  const formatTrack = ({ artist, name, url, date }) =>
    `- [${artist["#text"]} - ${name}](${url})${
      date ? ` (${date["#text"]})` : ""
    }`;
  const formatTracks = tracks => tracks.map(formatTrack).join("\n");

  bot.sendMessage(chatId, formatTracks(tracks), {
    parse_mode: "Markdown",
    disable_web_page_preview: true
  });
};

module.exports = recentTracks;
