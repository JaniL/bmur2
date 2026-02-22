import LastFm = require('lastfm-njs')

import {
  LastFmRecentTracksResponse,
  LastFmStore,
  LastFmTrack,
} from '../types/lastfm'

const { BMUR_LASTFM_APIKEY, BMUR_LASTFM_SECRET, BMUR_LASTFM_USERNAME } =
  process.env

if (!BMUR_LASTFM_APIKEY || !BMUR_LASTFM_SECRET || !BMUR_LASTFM_USERNAME) {
  throw new Error('Missing Last.fm environment variables')
}

const lfm = new LastFm({
  apiKey: BMUR_LASTFM_APIKEY,
  apiSecret: BMUR_LASTFM_SECRET,
  username: BMUR_LASTFM_USERNAME,
})

const requestOptions = {
  user: BMUR_LASTFM_USERNAME,
  limit: 15,
}

const emptyResponse: LastFmRecentTracksResponse = {
  track: [],
}

const store = (() => {
  let tracks: LastFmRecentTracksResponse = emptyResponse

  const setTracks = (
    incomingTracks: LastFmRecentTracksResponse,
  ): LastFmRecentTracksResponse => (tracks = incomingTracks)
  const getTracks = (): LastFmRecentTracksResponse => tracks

  return { setTracks, getTracks }
})()

const typedStore: LastFmStore = store

const normalizeTrack = (track: Partial<LastFmTrack>): LastFmTrack => ({
  artist: {
    '#text': track.artist?.['#text'] ?? '',
  },
  name: track.name ?? '',
  url: track.url ?? '',
  date: track.date,
})

const retrieveTracks = (): Promise<void> =>
  lfm
    .user_getRecentTracks(requestOptions)
    .then((response: LastFmRecentTracksResponse) => {
      const normalizedTracks = Array.isArray(response.track)
        ? response.track.map((track) => normalizeTrack(track))
        : []

      typedStore.setTracks({
        track: normalizedTracks,
      })
    })

setInterval(() => {
  void retrieveTracks()
}, 15 * 1000)
void retrieveTracks()

export default typedStore
