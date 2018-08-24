const LastFmNode = require("lastfm").LastFmNode;

const { BMUR_LASTFM_APIKEY, BMUR_LASTFM_SECRET } = process.env;

const lastfm = new LastFmNode({
  api_key: BMUR_LASTFM_APIKEY,
  secret: BMUR_LASTFM_SECRET,
  useragent: "bmurbot/v1.0"
});

const bmurStream = lastfm.stream("matlu_klusteri");
bmurStream.start();

module.exports = bmurStream;
