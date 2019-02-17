const reservationsApi = require('../integrations/reservations')

const upcomingReservations = bot => msg => {
  const chatId = msg.chat.id
  reservationsApi.retrieveUpcomingReservations()
    .then(res => {
      const formatReservations = reservations => {
        const formatLink = (name, id) => `[${name}](http://matlu.fi/ilotalo/index.php?page=res&id=${id})`
        return reservations.map(({ starts, name, association, id }) => `- ${starts}: ${formatLink(name, id)}, ${association}`).join('\n')
      }

      bot.sendMessage(chatId, formatReservations(res), {
        parse_mode: 'Markdown',
        disable_web_page_preview: true
      })
    })
}

module.exports = upcomingReservations