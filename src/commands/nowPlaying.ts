import moment = require('moment')
import TelegramBot = require('node-telegram-bot-api')

import { LastFmStore, LastFmTrack } from '../types/lastfm'

const formatTrack = ({ artist, name, url, date }: LastFmTrack): string => {
  const formatDate = (trackDate: NonNullable<LastFmTrack['date']>): string =>
    moment.unix(Number(trackDate.uts)).fromNow()

  return `[${artist['#text']} - ${name}](${url})${date ? ` (${formatDate(date)})` : ''}`
}

const nowPlaying =
  (bot: TelegramBot, lastfm: LastFmStore) =>
  (msg: TelegramBot.Message): void => {
    const chatId = msg.chat.id
    const [currentTrack] = lastfm.getTracks().track

    if (!currentTrack) {
      void bot.sendMessage(chatId, 'No recent tracks available right now.')
      return
    }

    void bot.sendMessage(chatId, formatTrack(currentTrack), {
      parse_mode: 'Markdown',
      disable_web_page_preview: true,
    })
  }

export default nowPlaying
