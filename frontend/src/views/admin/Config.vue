<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap)
      v-flex
        v-card
          v-card-title.headline リクエスト締め切り設定
          v-card-text
            span 企業
            date-time-picker(v-model="deadlines.day0" @input="updateDeadline(0, $event)")
            span 1日目
            date-time-picker(v-model="deadlines.day1" @input="updateDeadline(1, $event)")
            span 2日目
            date-time-picker(v-model="deadlines.day2" @input="updateDeadline(2, $event)")

</template>

<script>
import DateTimePicker from '../../components/DateTimePicker'
import gql from 'graphql-tag'

const getDeadlines = gql`
  query {
    day0: deadline(day: 0)
    day1: deadline(day: 1)
    day2: deadline(day: 2)
  }
`

const setDeadline = gql`
  mutation ($day: Int!, $time: Time!) {
    setDeadline(day: $day, time: $time)
  }
`

export default {
  name: 'Config',
  components: {
    DateTimePicker
  },
  data: function () {
    return {
      deadlines: {
        day0: null,
        day1: null,
        day2: null
      }
    }
  },
  apollo: {
    deadlines: {
      query: getDeadlines,
      fetchPolicy: 'network-only',
      update: data => data
    }
  },
  methods: {
    updateDeadline: async function (i, time) {
      try {
        await this.$apollo.mutate({
          mutation: setDeadline,
          variables: {
            day: i,
            time: time
          },
          update: (store, { data: { setDeadline } }) => {
            const data = store.readQuery({ query: getDeadlines })
            data['day' + i] = setDeadline
            store.writeQuery({ query: getDeadlines, data })
          }
        })
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
    }
  }
}
</script>
