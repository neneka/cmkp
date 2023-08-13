<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap)
      v-progress-linear(v-if="$apollo.queries.fetchData.loading" indeterminate)
      template(v-else)
        v-flex
          circle-detail-info(v-bind="fetchData.circle")
        v-flex(v-if="fetchData.circle.twitterId !== null" xs12 sm6 md6 lg4)
          v-card
            v-card-text
              timeline(:id="fetchData.circle.twitterId" sourceType="profile" :options="{ height: 400 }")
        v-flex(v-show="false")
          v-card
            v-card-title.headline サークルメモ
            v-card-text
              div(v-for="(memo, index) in fetchData.circleMemos" :key="memo.id")
                circle-memo(v-bind="memo" @deleted="onMemoDeleted")
                v-divider(v-if="index + 1 < fetchData.circleMemos.length")
            v-card-actions
              v-spacer
              v-dialog(v-model="dialog" persistent)
                v-btn(slot="activator" depressed color="primary") メモを書く
                v-card
                  v-card-title.headline メモを作成
                  v-card-text
                    | メモは全員に公開されます。
                    v-form(v-model="dialogValid")
                      v-textarea(label="内容" v-model="newMemo" :rules="[v => !!v || '内容を入力してください']" required)
                  v-card-actions
                    v-spacer
                    v-btn(flat @click.stop="dialog = false; newMemo = ''") キャンセル
                    v-btn(flat :disabled="!dialogValid || sending" :loading="sending" @click="createMemo") 作成
    v-btn(block color="success" :disabled="isDeadlineOver" :to="`/my-requests/create/${cid}`") リクエストを作成 {{ isDeadlineOver ? '(締め切りました)' : '' }}
    v-card
      v-list(three-line)
        v-list-tile(v-for="item in fetchData.circle.requestedItems", :key="item.id")
          v-list-tile-content
            v-list-tile-title {{ item.name }}
            v-list-tile-sub-title {{ priceString(item.price) }} × 計{{ requestedNum(item.requests) }}個
            v-list-tile-sub-title
              span(
                v-for="(request, idx) in item.requests",
                :key="request.id"
              )
                span
                  router-link(:to="`/planning/users/${request.userId}`") {{ request.user.displayName }}
                  | ({{ request.num }})
                template(v-if="idx !== item.requests.length - 1") ,&nbsp;
</template>

<script>
import gql from 'graphql-tag'
import CircleDetailInfo from '../components/CircleDetailInfo'
import CircleMemo from '../components/CircleMemo'
import { Timeline } from 'vue-tweet-embed'

const getData = gql`
  query ($cid: Int!) {
    circle(id: $cid) {
      id
      name
      author
      hall
      day
      block
      space
      locationType
      genre
      pixivId
      twitterId
      niconicoId
      website
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
    circleMemos(circleId: $cid) {
      id
      userId
      user {
        displayName
      }
      content
      createdAt
      updatedAt
    },
    deadlines {
      day
      over
    }
  }
`

const createCircleMemo = gql`
  mutation ($cid: Int!, $content: String!) {
    postCircleMemo(circleId: $cid, content: $content) {
      id
      userId
      user {
        displayName
      }
      content
      createdAt
      updatedAt
    }
  }
`

export default {
  name: 'CircleInfo',
  components: {
    CircleDetailInfo,
    CircleMemo,
    Timeline
  },
  props: {
    cid: {
      type: Number,
      required: true
    }
  },
  data: function () {
    return {
      fetchData: {
        circle: {
          day: null
        },
        circleMemos: [],
        deadlines: []
      },
      sending: false,
      dialog: false,
      dialogValid: false,
      newMemo: ''
    }
  },
  apollo: {
    fetchData: {
      query: getData,
      fetchPolicy: 'cache-and-network',
      variables: function () {
        return {
          cid: this.cid
        }
      },
      update: data => data
    }
  },
  computed: {
    isDeadlineOver () {
      for (let deadline of this.fetchData.deadlines) {
        if (deadline.day === this.fetchData.circle.day) {
          return deadline.over
        }
      }
      return false
    }
  },
  methods: {
    priceString: (p) => (p >= 0 ? `${p}円` : '価格未登録'),
    requestedNum: (requests) => requests.reduce((x, y) => x + y.num, 0),
    async createMemo () {
      this.sending = true
      try {
        await this.$apollo.mutate({
          mutation: createCircleMemo,
          variables: {
            cid: this.cid,
            content: this.newMemo
          },
          update: (store, { data: { postCircleMemo } }) => {
            const data = store.readQuery({ query: getData, variables: { cid: this.cid } })
            data.circleMemos.unshift(postCircleMemo)
            store.writeQuery({ query: getData, variables: { cid: this.cid }, data })
          }
        })
        this.dialog = false
        this.newMemo = ''
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    },
    onMemoDeleted (id) {
      let i = 0
      while (this.fetchData.circleMemos[i].id !== id) i++
      this.fetchData.circleMemos.splice(i, 1)
    }
  }
}
</script>
