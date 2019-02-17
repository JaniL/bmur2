require('dotenv').config()
const TelegramBot = require('node-telegram-bot-api')

const lastfm = require('./integrations/lastfm')
const nowPlaying = require('./commands/nowPlaying')
const recentTracks = require('./commands/recentTracks')
const upcomingReservations = require('./commands/upcomingReservations')

const { BMUR_TOKEN } = process.env

const token = BMUR_TOKEN
const bot = new TelegramBot(token, { polling: true })

bot.onText(/^\/np/, nowPlaying(bot, lastfm))
bot.onText(/^\/recent/, recentTracks(bot, lastfm))
bot.onText(/^\/reservations/, upcomingReservations(bot))
