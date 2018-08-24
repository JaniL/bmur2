const recentTracks = (bot, lastfm) => msg => {
  const chatId = msg.chat.id;

  const { track: tracks } = lastfm.getTracks();

  const formatTrack = ({ artist, name }) => `- ${artist["#text"]} - ${name}`;
  const formatTracks = tracks => tracks.map(formatTrack).join("\n");

  bot.sendMessage(chatId, formatTracks(tracks));
};

module.exports = recentTracks;
