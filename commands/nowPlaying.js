const nowPlaying = (bot, lastfm) => msg => {
  const chatId = msg.chat.id;
  const {
    track: [nowPlaying]
  } = lastfm.getTracks();
  console.log(nowPlaying);

  const formatTrack = ({ artist, name, url }) =>
    `[${artist["#text"]} - ${name}](${url})`;
  bot.sendMessage(chatId, formatTrack(nowPlaying), {
    parse_mode: "Markdown",
    disable_web_page_preview: true
  });
};

module.exports = nowPlaying;
