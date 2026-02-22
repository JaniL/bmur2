import 'dotenv/config'
import TelegramBot = require('node-telegram-bot-api')

import nowPlaying from './commands/nowPlaying'
import recentTracks from './commands/recentTracks'
import lastfm from './integrations/lastfm'

const token = process.env.BMUR_TOKEN

if (!token) {
  throw new Error('Missing BMUR_TOKEN environment variable')
}

const bot = new TelegramBot(token, { polling: true })

bot.onText(/^\/np/, nowPlaying(bot, lastfm))
bot.onText(/^\/recent/, recentTracks(bot, lastfm))
