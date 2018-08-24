const nowPlaying = (bot, lastfm) => msg => {
  const chatId = msg.chat.id;
  const {
    track: [nowPlaying]
  } = lastfm.getTracks();
  console.log(nowPlaying);

  const formatTrack = ({ artist, name }) => `${artist["#text"]} - ${name}`;
  bot.sendMessage(chatId, formatTrack(nowPlaying));
};

module.exports = nowPlaying;
