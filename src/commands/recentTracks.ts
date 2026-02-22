import moment = require('moment')
import TelegramBot = require('node-telegram-bot-api')

import { LastFmStore, LastFmTrack } from '../types/lastfm'

const formatTrack = ({ artist, name, url, date }: LastFmTrack): string => {
  const formatDate = (trackDate: NonNullable<LastFmTrack['date']>): string =>
    moment.unix(Number(trackDate.uts)).fromNow()

  return `- [${artist['#text']} - ${name}](${url})${date ? ` (${formatDate(date)})` : ''}`
}

const formatTracks = (tracks: LastFmTrack[]): string =>
  tracks.map(formatTrack).join('\n')

const recentTracks =
  (bot: TelegramBot, lastfm: LastFmStore) =>
  (msg: TelegramBot.Message): void => {
    const chatId = msg.chat.id
    const tracks = lastfm.getTracks().track

    if (tracks.length === 0) {
      void bot.sendMessage(chatId, 'No recent tracks available right now.')
      return
    }

    void bot.sendMessage(chatId, formatTracks(tracks), {
      parse_mode: 'Markdown',
      disable_web_page_preview: true,
    })
  }

export default recentTracks
