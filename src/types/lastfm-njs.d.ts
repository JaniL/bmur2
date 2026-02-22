declare module 'lastfm-njs' {
  interface LastFmOptions {
    apiKey: string
    apiSecret: string
    username: string
  }

  interface GetRecentTracksOptions {
    user: string
    limit: number
  }

  interface LastFmRecentTrack {
    artist: {
      '#text': string
    }
    name: string
    url: string
    date?: {
      uts: string
    }
  }

  interface LastFmRecentTracksResponse {
    track: LastFmRecentTrack[]
  }

  class LastFm {
    constructor(options: LastFmOptions)

    user_getRecentTracks(
      options: GetRecentTracksOptions,
    ): Promise<LastFmRecentTracksResponse>
  }

  export = LastFm
}
