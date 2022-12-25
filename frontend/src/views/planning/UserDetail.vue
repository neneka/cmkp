<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap v-if="!$apollo.queries.fetchData.loading")
      v-flex(d-block sm12 md4)
        v-layout(row wrap)
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline {{ user.displayName }} (@{{ user.name }})
              v-card-text
                v-chip(v-if="user.role === 'PLANNER'" color="green" text-color="white" small) プランナー
                v-chip(v-else-if="user.role === 'ADMIN'" color="red" text-color="white" small) 管理人
                v-chip(v-for="day in user.entries" :key="day" color="primary" text-color="white" small) {{ day }}日目
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline リクエスト備考
              v-card-text
                div(v-for="note in notes" :key="note.id")
                  span.caption {{ formatDatetime(note.updatedAt) }}
                  br
                  span.body-1.user-content-text(v-text="note.content" v-linkified)
      v-flex(d-flex sm12 md8)
        v-layout(row wrap)
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline 希望順位
              v-container(fluid grid-list-xs)
                v-layout(row wrap)
                  v-flex(xs12 sm4 md4 lg4)
                    circle-priority-list(title="企業" :circles="priorities[0]")
                  v-flex(v-for="i in 2" xs12 sm4 md4 lg4 :key="i")
                    circle-priority-list(:title="`${i}日目`" :circles="priorities[i]")
          v-flex(d-flex xs12)
            v-card
              v-card-title
                span.headline リクエストリスト
              v-container(fluid grid-list-xs)
                v-btn(block depressed color="primary" @click="addItemDialog.open = true" append) リクエスト追加
                v-radio-group(v-model="filter.day" row)
                  v-radio(:label="`企業(${requestedCircleCounts[0]})`" :value="0")
                  v-radio(v-for="i in 2" :label="`${i}日目(${requestedCircleCounts[i]})`" :value="i" :key="i")
                v-layout(row wrap)
                  v-flex(xs12 sm12 md6 lg4 v-for="circle in filteredRequests" :key="circle.id")
                    v-card
                      v-card-title.headline.lighten-4(:class="[{'orange': circle.locationType === 1}, {'red': circle.locationType === 2}, {'green': circle.locationType === 0}]")
                        router-link(:to="`/circles/${circle.id}`" style="text-decoration: none;") {{ circle.locationString }} {{ circle.name }} ({{ circle.author }})
                      v-divider
                      v-list
                        v-list-tile(v-for="item in circle.items" :key="item.id" @click="onItemClicked(circle, item)")
                          v-list-tile-content
                            v-list-tile-title {{ item.name }}
                            v-list-tile-sub-title {{ priceString(item.price) }} × {{ item.request.num }}個
    v-dialog(v-model="editItemDialog.open" width=500 persistent)
      v-card
        v-card-title.headline {{ editItemDialog.origItem.name }}
        v-card-text
          v-form(v-model="editItemDialog.valid")
            v-text-field(v-model.number="editItemDialog.price" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="editItemDialog.sending")
            v-text-field(v-model.number="editItemDialog.num" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="editItemDialog.sending")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="editItemDialog.open = false" :disabled="editItemDialog.sending") キャンセル
          v-btn(flat color="red" @click.native="deleteRequest" :disabled="editItemDialog.sending" :loading="editItemDialog.deleting") 削除
          v-btn(flat color="primary" @click.native="editRequest" :disabled="editItemDialog.sending || !editItemDialog.valid" :loading="editItemDialog.editing") 修正
    v-dialog(v-model="addItemDialog.open" persistent)
      v-card
        v-toolbar(card)
          v-btn(icon @click="addItemDialog.open = false")
            v-icon close
          v-toolbar-title {{ user.displayName }}のリクエストを追加
          v-spacer
          v-toolbar-items
            v-btn(color="primary" :disabled="!addItemDialog.valid || addItemDialog.sending" :loading="addItemDialog.sending" @click="createRequest") 登録
        v-card-text
          v-form(v-model="addItemDialog.valid")
            v-text-field(v-model="addItemDialog.query" label="サークル名または作家名を入力" maxLength="100" counter persistent-hint :loading="$apollo.queries.searchCircles.loading")
            v-select(v-model="addItemDialog.circle" hide-no-data hide-selected :item-text="v => `${v.locationString} ${v.name} ${v.author}`" item-value="id" label="サークル選択" return-object placeholder="サークル名または作家名を選択" :items="searchCircles" :loading="$apollo.queries.searchCircles.loading" required :rules="[rules.required]")
            v-select(v-model="addItemDialog.selectedItem" :items="selectableCircleItems" label="商品名を選択" hint="希望の商品が無い場合は新規登録を選んでください" persistent-hint return-object single-line item-text="name" item-value="id" :loading="$apollo.queries.circleItems.loading")
            v-text-field(v-model="addItemDialog.name" label="商品名を入力" hint="曖昧な商品名を入力しないでください。また複数の商品を一つにまとめて登録しないでください。(OK：新刊A, NG：新刊AとB)" required :rules="[rules.required]" maxLength="100" counter persistent-hint :disabled="addItemDialog.selectedItem == null || !isNewItem")
            v-text-field(v-model.number="addItemDialog.price" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="addItemDialog.selectedItem == null")
            v-text-field(v-model.number="addItemDialog.num" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="addItemDialog.selectedItem == null")

</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'
import updateItemPrice from '../../gql/updateItemPrice.gql'
import changeRequestNum from '../../gql/changeRequestNum.gql'
import deleteRequest from '../../gql/deleteRequest.gql'
import createItem from '../../gql/createItem.graphql'
import createRequest from '../../gql/createRequest.gql'
import CirclePriorityList from '../../components/CirclePriorityList'

const getData = gql`
  query ($uid: Int!) {
    user(id: $uid) {
      id
      name
      displayName
      role
      entries
    }
    requestNotes(userId: $uid) {
      id
      content
      updatedAt
    }
    userRequestedCircles(userId: $uid) {
      id
      name
      author
      day
      locationString
      locationType
      items: requestedItems(userId: $uid) {
        id
        name
        price
        request: userRequest(userId: $uid) {
          id
          num
        }
      }
    }
    priority0: circlePriority(userId: $uid, day: 0) {
      circles {
        id
        name
      }
    }
    priority1: circlePriority(userId: $uid, day: 1) {
      circles {
        id
        name
      }
    }
    priority2: circlePriority(userId: $uid, day: 2) {
      circles {
        id
        name
      }
    }
    priority3: circlePriority(userId: $uid, day: 3) {
      circles {
        id
        name
      }
    }
    priority4: circlePriority(userId: $uid, day: 4) {
      circles {
        id
        name
      }
    }
  }
`

const searchCircles = gql`
  query ($q: String!) {
    searchCircles: circles(q: $q) {
      id
      name
      author
      locationString(day: true)
    }
  }
`

const getCircleItems = gql`
  query ($cid: Int!) {
    circleItems: items(circleId: $cid) {
      id
      name
      price
    }
  }
`

export default {
  name: 'UserDetail',
  components: {
    CirclePriorityList
  },
  props: {
    userId: {
      type: Number,
      required: true
    }
  },
  data: function () {
    return {
      fetchData: {
        user: {},
        requestNotes: [],
        userRequestedCircles: [],
        priority0: {
          circles: []
        },
        priority1: {
          circles: []
        },
        priority2: {
          circles: []
        },
        priority3: {
          circles: []
        },
        priority4: {
          circles: []
        }
      },
      filter: {
        day: 0
      },
      editItemDialog: {
        origItem: {
          name: ''
        },
        origCircle: null,
        open: false,
        sending: false,
        deleting: false,
        editing: false,
        price: -1,
        num: 1,
        valid: false
      },
      addItemDialog: {
        open: false,
        sending: false,
        valid: false,
        circle: null,
        selectedItem: null,
        query: '',
        name: '',
        price: '',
        num: 1
      },
      searchCircles: [],
      circleItems: [],
      rules: {
        required: value => !!value || '必須項目です'
      }
    }
  },
  computed: {
    user: function () {
      return this.fetchData.user
    },
    notes: function () {
      return this.fetchData.requestNotes
    },
    priorities: function () {
      return [this.fetchData.priority0.circles,
        this.fetchData.priority1.circles,
        this.fetchData.priority2.circles,
        this.fetchData.priority3.circles,
        this.fetchData.priority4.circles]
    },
    circles: function () {
      return this.fetchData.userRequestedCircles
    },
    filteredRequests: function () {
      return this.circles.filter(v => v.day === this.filter.day)
    },
    requestedCircleCounts: function () {
      return this.circles.reduce((x, y) => {
        x[y.day]++
        return x
      }, [0, 0, 0, 0, 0])
    },
    selectableCircleItems: function () {
      return [{
        name: '新規登録',
        id: null
      }, ...this.circleItems]
    },
    isNewItem: function () {
      return this.addItemDialog.selectedItem !== null && this.addItemDialog.selectedItem.id === null
    }
  },
  watch: {
    'addItemDialog.selectedItem': function () {
      if (!this.addItemDialog.selectedItem) return
      if (this.isNewItem) {
        this.addItemDialog.name = ''
        this.addItemDialog.price = ''
        this.addItemDialog.num = 1
      } else {
        this.addItemDialog.name = this.addItemDialog.selectedItem.name
        this.addItemDialog.price = this.addItemDialog.selectedItem.price >= 0 ? this.addItemDialog.selectedItem.price : ''
        this.addItemDialog.num = 1
      }
    },
    'circleItems': function () {
      this.addItemDialog.selectedItem = null
    }
  },
  apollo: {
    fetchData: {
      query: getData,
      fetchPolicy: 'cache-and-network',
      variables: function () {
        return {
          uid: this.userId
        }
      },
      update: data => data
    },
    searchCircles: {
      query: searchCircles,
      variables: function () {
        return {
          q: this.addItemDialog.query || ''
        }
      },
      debounce: 500,
      skip: function () {
        return this.addItemDialog.query === '' || this.addItemDialog.circle != null
      }
    },
    circleItems: {
      query: getCircleItems,
      variables: function () {
        return {
          cid: this.addItemDialog.circle.id
        }
      },
      skip: function () {
        return this.addItemDialog.circle == null
      }
    }
  },
  methods: {
    priceString: p => p >= 0 ? `${p}円` : '価格未登録',
    formatDatetime: dt => dayjs(dt).fromNow(),
    onItemClicked: function (circle, item) {
      this.editItemDialog.origCircle = circle
      this.editItemDialog.origItem = item
      this.editItemDialog.price = item.price === -1 ? '' : item.price
      this.editItemDialog.num = item.request.num
      this.editItemDialog.open = true
    },
    createRequest: async function () {
      this.addItemDialog.sending = true
      try {
        let itemId
        if (!this.isNewItem) {
          itemId = this.addItemDialog.selectedItem.id
          if (this.addItemDialog.price !== this.addItemDialog.selectedItem.price) {
            await this.$apollo.mutate({
              mutation: updateItemPrice,
              variables: {
                id: itemId,
                price: this.addItemDialog.price !== '' ? this.addItemDialog.price : -1
              }
            })
          }
        } else {
          itemId = (await this.$apollo.mutate({
            mutation: createItem,
            variables: {
              cid: this.addItemDialog.circle.id,
              name: this.addItemDialog.name,
              price: this.addItemDialog.price !== '' ? this.addItemDialog.price : -1
            }
          })).data.createItem.id
        }

        await this.$apollo.mutate({
          mutation: createRequest,
          variables: {
            userId: this.userId,
            itemId: itemId,
            num: this.addItemDialog.num
          }
        })
        this.$apollo.queries.fetchData.refetch()
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.addItemDialog.circle = null
      this.addItemDialog.selectedItem = null
      this.addItemDialog.open = false
      this.addItemDialog.sending = false
    },
    editRequest: async function () {
      this.editItemDialog.sending = true
      this.editItemDialog.editing = true
      try {
        const item = this.editItemDialog.origItem
        const req = item.request

        const p = this.editItemDialog.price === '' ? -1 : this.editItemDialog.price
        if (req.num !== this.editItemDialog.num) {
          await this.$apollo.mutate({
            mutation: changeRequestNum,
            variables: {
              id: req.id,
              num: this.editItemDialog.num
            }
          })
        }
        if (item.price !== p) {
          await this.$apollo.mutate({
            mutation: updateItemPrice,
            variables: {
              id: item.id,
              price: p
            }
          })
        }

        req.num = this.editItemDialog.num
        item.price = p
        this.editItemDialog.open = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.editItemDialog.editing = false
      this.editItemDialog.sending = false
    },
    deleteRequest: async function () {
      this.editItemDialog.sending = true
      this.editItemDialog.deleting = true
      try {
        await this.$apollo.mutate({
          mutation: deleteRequest,
          variables: {
            id: this.editItemDialog.origItem.request.id
          }
        })

        const circle = this.editItemDialog.origCircle
        let i = 0
        while (circle.items[i].id !== this.editItemDialog.origItem.id) i++
        circle.items.splice(i, 1)

        this.editItemDialog.open = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.editItemDialog.deleting = false
      this.editItemDialog.sending = false
    }
  }
}
</script>
