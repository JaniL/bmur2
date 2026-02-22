export interface LastFmArtist {
  '#text': string
}

export interface LastFmTrackDate {
  uts: string
}

export interface LastFmTrack {
  artist: LastFmArtist
  name: string
  url: string
  date?: LastFmTrackDate
}

export interface LastFmRecentTracksResponse {
  track: LastFmTrack[]
}

export interface LastFmStore {
  getTracks: () => LastFmRecentTracksResponse
  setTracks: (
    incomingTracks: LastFmRecentTracksResponse,
  ) => LastFmRecentTracksResponse
}
