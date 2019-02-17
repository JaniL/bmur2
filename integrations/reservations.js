const axios = require('axios')

const reservationsApi = axios.create({
  baseURL: 'http://188.166.126.9:3000'
})

const retrieveUpcomingReservations = () =>
  reservationsApi.get('/reservations/upcoming')
    .then(res => res.data)

module.exports = { retrieveUpcomingReservations }