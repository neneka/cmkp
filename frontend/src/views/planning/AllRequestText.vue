<template lang="pug">
v-container(fluid, grid-list-md)
  v-layout(column)
    v-flex(xs12)
      template(v-for="section in all")
        details
          summary.headline {{ section.title }}
          v-card
            v-card-text
              v-btn(depressed, small, @click.stop="copy(section.text)") コピー
              v-btn(
                depressed,
                small,
                @click.stop="download(section.title, section.text)"
              ) ダウンロード
              codeblock
                pre {{ section.text }}
</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'

const getData = gql`
  query {
    requestedCircles {
      id
      name
      author
      hall
      block
      day
      locationString
      locationType
      prioritized {
        userId
        user {
          displayName
        }
        rank
      }
      requestedItems {
        id
        name
        price
        requests {
          id
          userId
          user {
            displayName
          }
          num
        }
      }
    }
  }
`

const KANA_123 = [
  'ア',
  'イ',
  'ウ',
  'エ',
  'オ',
  'カ',
  'キ',
  'ク',
  'ケ',
  'コ',
  'サ',
  'シ'
]

export default {
  name: 'AllRequestText',
  data: function () {
    return {
      circles: []
    }
  },
  apollo: {
    circles: {
      query: getData,
      fetchPolicy: 'cache-and-network',
      update: (data) =>
        data.requestedCircles.sort((a, b) =>
          a.locationString.localeCompare(b.locationString, 'ja', {
            numeric: true
          })
        )
    }
  },
  computed: {
    company: function () {
      return this.circles.filter((circle) => circle.day === 0)
    },
    day1_east_123: function () {
      return this.circles.filter(
        (circle) =>
          circle.day === 1 &&
          this.isEast(circle.hall) &&
          this.isBlock123(circle.block)
      )
    },
    day1_east_456: function () {
      return this.circles.filter(
        (circle) =>
          circle.day === 1 &&
          this.isEast(circle.hall) &&
          !this.isBlock123(circle.block)
      )
    },
    day1_west: function () {
      return this.circles.filter(
        (circle) => circle.day === 1 && !this.isEast(circle.hall)
      )
    },
    day2_east_123: function () {
      return this.circles.filter(
        (circle) =>
          circle.day === 2 &&
          this.isEast(circle.hall) &&
          this.isBlock123(circle.block)
      )
    },
    day2_east_456: function () {
      return this.circles.filter(
        (circle) =>
          circle.day === 2 &&
          this.isEast(circle.hall) &&
          !this.isBlock123(circle.block)
      )
    },
    day2_west: function () {
      return this.circles.filter(
        (circle) => circle.day === 2 && !this.isEast(circle.hall)
      )
    },
    all: function () {
      return [
        ['企業', this.company],
        ['1日目 東 123ホール', this.day1_east_123],
        ['1日目 東 456ホール', this.day1_east_456],
        ['1日目 西', this.day1_west],
        ['2日目 東 123ホール', this.day2_east_123],
        ['2日目 東 456ホール', this.day2_east_456],
        ['2日目 西', this.day2_west]
      ].map(([title, circles]) => ({
        title,
        text: circles.map((circle) => this.genText(circle)).join('\n\n')
      }))
    }
  },
  methods: {
    datetimeString: function (dt) {
      return dayjs(dt).fromNow()
    },
    replaceFullToHalf: function (str) {
      return str.replace(/[！-～]/g, function (s) {
        return String.fromCharCode(s.charCodeAt(0) - 0xfee0)
      })
    },
    isEast: function (s) {
      return s === '東'
    },
    isBlock123: function (s) {
      return (
        this.replaceFullToHalf(s).match(/^[A-Z]$/) ||
        KANA_123.find((k) => k === s)
      )
    },
    genText: function (circle) {
      return `${circle.locationString} ${circle.name} (${circle.author})${
        circle.prioritized.length > 0 ? '\n' : ''
      }${circle.prioritized
        .map((prio) => `第${prio.rank}希望: ${prio.user.displayName}`)
        .join(', ')}\n${circle.requestedItems
        .map(
          (item) =>
            `${item.name}\n${
              item.price >= 0 ? item.price : '価格未登録'
            }円 x 計${item.requests
              .map((request) => request.num)
              .reduce((a, b) => a + b)}個\n${item.requests
              .map((request) => `${request.user.displayName} (${request.num})`)
              .join(', ')}`
        )
        .join('\n')}`
    },
    copy: function (text) {
      navigator.clipboard.writeText(text)
    },
    download: function (title, text) {
      try {
        const blob = new Blob([text], {
          type: 'text/plain'
        })
        const url = URL.createObjectURL(blob)
        const button = document.createElement('a')
        button.type = 'button'
        button.download = `cmkp-${title}.txt`
        button.href = url
        button.click()
      } catch (error) {
        console.error(error)
        alert('生成に失敗しました')
      }
    }
  }
}
</script>
