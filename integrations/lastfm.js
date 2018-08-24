const lastfm = require("lastfm-njs");

const {
  BMUR_LASTFM_APIKEY,
  BMUR_LASTFM_SECRET,
  BMUR_LASTFM_USERNAME
} = process.env;

const lfm = new lastfm({
  apiKey: BMUR_LASTFM_APIKEY,
  apiSecret: BMUR_LASTFM_SECRET,
  username: BMUR_LASTFM_USERNAME
});

const requestOptions = {
  user: "matlu_klusteri",
  limit: 15
};

const store = (() => {
  let tracks = {};

  const setTracks = incomingTracks => (tracks = incomingTracks);
  const getTracks = () => tracks;
  return { setTracks, getTracks };
})();

const retrieveTracks = () =>
  lfm
    .user_getRecentTracks(requestOptions)
    .then(tracks => store.setTracks(tracks));

setInterval(retrieveTracks, 15 * 1000);
retrieveTracks();

module.exports = store;
