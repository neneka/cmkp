<template lang="pug">
  v-container(fluid grid-list-md)
    v-card
      v-card-text
        | 右下の+ボタンまたはサークル別の商品追加ボタンからリクエストを作成できます。
        br
        | 登録商品を押すと、登録情報を修正・削除できます。サークル名を押すとサークル詳細を確認できます。
        v-select(label="日程" :items="daySelectItems" v-model="selectedDay")
        p 合計：{{ circleCount }}サークル, {{ totalCosts }}
        v-btn(color="primary" depressed block :disabled="isSelectedDayDeadlineOver" @click="openEditPriorityDialog") {{selectedDay !== '全日' ? `希望優先順位を設定する` : `日程を選択してください`}}
    v-card
      v-card-text
        v-btn(color="primary" depressed block :disabled="!requestedCircles" @click="downloadRequests") リクエストのJSONをダウンロードする

    v-layout(row wrap)
      v-flex(xs12 sm12 md6 lg4 v-for="circle in filteredRequests" :key="circle.id")
        v-card
          v-card-title.headline.lighten-4(:class="[{'orange': circle.locationType === 1}, {'red': circle.locationType === 2}, {'green': circle.locationType === 0}]")
            router-link(:to="`/circles/${circle.id}`" style="text-decoration: none;") {{ circle.locationString }} {{ circle.name }}
          v-list
            v-list-tile(v-for="request in circle.requests" :key="request.id" @click="openEditDialog(request)")
              v-list-tile-content
                v-list-tile-title {{ request.item.name }}
                v-list-tile-sub-title {{ request.item.price > -1 ? request.item.price + '円' : '価格未定' }} × {{ request.num }}個
          v-card-actions
            p 合計：{{ sumCosts(circle.requests) }}
            v-spacer
            v-btn(depressed :to="`/my-requests/create/${circle.id}`" :disabled="isDeadlineOver(circle.day)") 商品追加

    v-btn(fixed dark fab bottom right color="blue darken-2" to="/circles")
      v-icon add

    v-dialog(v-model="dialog" width=500 persistent)
      v-card
        v-card-title.headline {{ editTarget != null ? editTarget.item.name : ''}}
        v-card-text
          v-form(v-model="valid")
            v-text-field(v-model.number="editPrice" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="sending")
            v-text-field(v-model.number="editNum" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="!numEditable || sending")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="dialog = false" :disabled="sending") キャンセル
          v-btn(flat color="red" @click.native="deleteRequest" :disabled="sending || !numEditable" :loading="deleting") 削除
          v-btn(flat color="primary" @click.native="editRequest" :disabled="sending || !valid" :loading="editing") 修正

    v-dialog(v-model="priorityDialog" width=500 persistent)
      v-card
        v-card-title.headline {{ selectedDay }}の希望優先順位
        v-card-text 上位5つまで設定できます。
        v-list
          v-list-tile(v-for="(p, i) of editPriorities" :key="p.id")
            v-list-tile-content
              v-list-tile-title 第{{ i + 1 }}希望 {{ p.name }}
            v-list-tile-action
              v-btn(icon @click="removePriorityCircle(i)")
                v-icon clear
        v-card-text
          v-select(v-model="editPrioritySelectCircle" :items="prioritySelectItems" placeholder="ここから選択" return-object item-text="name" item-value="id" append-outer-icon="add" @click:append-outer="appendPriorityCircle" :disabled="editPriorities.length >= 5")

        v-card-actions
          v-spacer
          v-btn(flat @click.native="priorityDialog = false" :disabled="sending") キャンセル
          v-btn(flat color="primary" :disabled="sending" :loading="sending" @click.native="updatePriority") OK

</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'
import updateItemPrice from '../gql/updateItemPrice.gql'
import editRequest from '../gql/changeRequestNum.gql'
import deleteRequest from '../gql/deleteRequest.gql'

const getData = gql`
  query {
    myRequestedCircles {
      id
      name
      author
      day
      locationString
      locationType
    }
    myRequests {
      id
      item {
        id
        circleId
        name
        price
      }
      num
    }
    priorities0: myCirclePriorityIds(day: 0)
    priorities1: myCirclePriorityIds(day: 1)
    priorities2: myCirclePriorityIds(day: 2)
    priorities3: myCirclePriorityIds(day: 3)
    priorities4: myCirclePriorityIds(day: 4)
    day0: deadline(day: 0)
    day1: deadline(day: 1)
    day2: deadline(day: 2)
  }
`

const updateCirclePriority = gql`
  mutation ($day: Int!, $circles: [Int!]!) {
    setCirclePriorities(day: $day, circleIds: $circles) {
      day
      priorities
    }
  }
`

export default {
  name: 'MyRequests',
  data: function () {
    return {
      fetchData: {
        myRequestedCircles: [],
        myRequests: [],
        priorities0: [],
        priorities1: [],
        priorities2: [],
        priorities3: [],
        priorities4: [],
        day0: null,
        day1: null,
        day2: null
      },
      selectedDay: '全日',
      daySelectItems: ['企業', '１日目', '２日目', '全日'],
      requestCircleMap: new Map(),
      dialog: false,
      priorityDialog: false,
      sending: false,
      deleting: false,
      editing: false,
      editTarget: null,
      editTargetCircle: null,
      editNum: 1,
      editPrice: -1,
      editPriorities: [],
      editPrioritySelectCircle: null,
      rules: {
        required: value => !!value || '必須項目です'
      },
      valid: false
    }
  },
  computed: {
    requestedCircles: function () {
      const result = []
      this.requestCircleMap.clear()
      for (const circle of this.fetchData.myRequestedCircles) {
        const c = Object.assign({}, circle)
        result.push(c)
        this.requestCircleMap.set(c.id, c)
      }
      for (const req of this.fetchData.myRequests) {
        const c = this.requestCircleMap.get(req.item.circleId)
        if (!c.requests) {
          c.requests = []
        }
        c.requests.push(req)
      }
      return result.sort((a, b) => a.locationString.localeCompare(b.locationString, 'ja', { numeric: true }))
    },
    numEditable: function () {
      if (this.editTarget == null) {
        return false
      }
      return !this.isDeadlineOver(this.requestCircleMap.get(this.editTarget.item.circleId).day)
    },
    filteredRequests: function () {
      const day = this.selectedDayNum
      if (day == null) return this.requestedCircles
      return this.requestedCircles.filter(v => v.day === day)
    },
    totalCosts: function () {
      let unknown = false
      let costs = 0
      for (const circle of this.filteredRequests) {
        for (const req of circle.requests) {
          if (req.item.price === -1) {
            unknown = true
          } else {
            costs += req.item.price * req.num
          }
        }
      }

      if (unknown) {
        return `${costs}円 + α`
      } else {
        return `${costs}円`
      }
    },
    circleCount: function () {
      return this.filteredRequests.length
    },
    prioritySelectItems: function () {
      return this.filteredRequests.filter(v => this.editPriorities.every(w => v.id !== w.id))
    },
    selectedDayNum: function () {
      switch (this.selectedDay) {
        case '企業':
          return 0
        case '１日目':
          return 1
        case '２日目':
          return 2
        default:
          return null
      }
    },
    isSelectedDayDeadlineOver: function () {
      return this.isDeadlineOver(this.selectedDayNum)
    }
  },
  apollo: {
    fetchData: {
      query: getData,
      fetchPolicy: 'cache-and-network',
      update: data => data
    }
  },
  created: async function () {
    if (this.$route.query.day) {
      switch (this.$route.query.day) {
        case '0':
          this.selectedDay = '企業'
          break
        case '1':
          this.selectedDay = '１日目'
          break
        case '2':
          this.selectedDay = '２日目'
          break
      }
    }
  },
  methods: {
    isDeadlineOver: function (day) {
      switch (day) {
        case 0:
          return dayjs(this.fetchData.day0).isBefore(dayjs())
        case 1:
          return dayjs(this.fetchData.day1).isBefore(dayjs())
        case 2:
          return dayjs(this.fetchData.day2).isBefore(dayjs())
        default:
          return true
      }
    },
    sumCosts: function (requests) {
      let unknown = false
      let costs = 0
      for (const req of requests) {
        if (req.item.price === -1) {
          unknown = true
        } else {
          costs += req.item.price * req.num
        }
      }

      if (unknown) {
        return `${costs}円 + α`
      } else {
        return `${costs}円`
      }
    },
    openEditDialog: function (req) {
      this.editTarget = req
      this.editNum = req.num
      this.editPrice = req.item.price === -1 ? '' : req.item.price
      this.dialog = true
    },
    editRequest: async function () {
      this.sending = true
      this.editing = true
      try {
        const p = this.editPrice === '' ? -1 : this.editPrice
        if (this.editTarget.num !== this.editNum) {
          await this.$apollo.mutate({
            mutation: editRequest,
            variables: {
              id: this.editTarget.id,
              num: this.editNum
            }
          })
        }
        if (this.editTarget.item.price !== p) {
          await this.$apollo.mutate({
            mutation: updateItemPrice,
            variables: {
              id: this.editTarget.item.id,
              price: p
            }
          })
        }

        this.editTarget.num = this.editNum
        this.editTarget.item.price = p
        this.dialog = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.editing = false
      this.sending = false
    },
    deleteRequest: async function () {
      this.sending = true
      this.deleting = true
      try {
        await this.$apollo.mutate({
          mutation: deleteRequest,
          variables: {
            id: this.editTarget.id
          }
        })
        const circle = this.requestCircleMap.get(this.editTarget.item.circleId)
        let i = 0
        while (circle.requests[i].id !== this.editTarget.id) i++
        circle.requests.splice(i, 1)

        this.dialog = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.deleting = false
      this.sending = false
    },
    openEditPriorityDialog: function () {
      let ids
      switch (this.selectedDayNum) {
        case 0:
          ids = this.fetchData.priorities0
          break
        case 1:
          ids = this.fetchData.priorities1
          break
        case 2:
          ids = this.fetchData.priorities2
          break
        case 3:
          ids = this.fetchData.priorities3
          break
        case 4:
          ids = this.fetchData.priorities4
          break
        default:
          return
      }

      this.editPriorities.splice(0)
      if (ids != null) {
        for (const id of ids) {
          const c = this.requestCircleMap.get(id)
          if (c) {
            this.editPriorities.push(c)
          }
        }
      }
      this.priorityDialog = true
    },
    appendPriorityCircle: function () {
      if (this.editPrioritySelectCircle == null) return
      this.editPriorities.push(this.editPrioritySelectCircle)
      this.editPrioritySelectCircle = null
    },
    removePriorityCircle: function (idx) {
      this.editPriorities.splice(idx, 1)
    },
    updatePriority: async function () {
      this.sending = true
      try {
        await this.$apollo.mutate({
          mutation: updateCirclePriority,
          variables: {
            day: this.selectedDayNum,
            circles: this.editPriorities.map(v => v.id)
          },
          update: (store, { data: { setCirclePriorities } }) => {
            const data = store.readQuery({ query: getData })
            data[`priorities${setCirclePriorities.day}`] = setCirclePriorities.priorities
            store.writeQuery({ query: getData, data })
          }
        })
        this.priorityDialog = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    },
    downloadRequests: function () {
      try {
        const data = this.fetchData.myRequests.map(({ item, num }) => {
          const circle = window.structuredClone(this.fetchData.myRequestedCircles.find((circle) => circle.id === item.circleId))
          delete circle.__typename
          delete circle.id
          return {
            name: item.name,
            price: item.price,
            num,
            circle
          }
        }).sort((a, b) => a.circle.locationString.localeCompare(b.circle.locationString, 'ja', { numeric: true }))
        console.log(data)
        const blob = new Blob([JSON.stringify(data, undefined, 2)], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const button = document.createElement('a')
        button.type = 'button'
        button.download = 'cmkp-C100-requests.json'
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
