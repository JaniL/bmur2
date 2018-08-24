const nowPlaying = (bot, lastfm) => {
  let nowPlaying = {};
  lastfm.on("nowPlaying", track => (nowPlaying = track));
  lastfm.on("lastPlayed", track => (nowPlaying = track));
  return msg => {
    const chatId = msg.chat.id;
    console.log(track);
    console.log(chatId);
    bot.sendMessage(chatId, "asd");
  };
};

module.exports = nowPlaying;
