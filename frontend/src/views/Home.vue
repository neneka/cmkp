<template lang="pug">
v-container(fluid, grid-list-md)
  v-layout(column)
    v-flex
      v-card
        v-card-text
          span ◎サイドメニュー(左上のボタン)の「マイリクエスト」または「サークル検索」から、買ってきて欲しい物のリクエスト等を行ってください。
          br
          span ◎リクエストをし終わったら、必ず「マイリクエスト」の「希望優先順位を設定する」から特に優先して買ってきて欲しいサークルを、上位５サークルまで日程毎に設定してください。
          br
          span ◎リクエストに関して管理人に伝えたいことがある場合は、メニューの「リクエスト備考」からリクエスト備考を書くことができます。
          br
          span ◎リクエスト受付締め切り後は、リクエストの登録・削除・個数の変更・優先順位設定は出来ません。何かあれば管理人に個別に連絡してください。
          br
          span ◎限数・搬入数やお品書きなどの情報があればリクエスト備考に書いてもらえると助かります。また、価格未定で登録した商品は価格が決まり次第、リクエストの修正から入力してください。
    v-flex
      v-card
        v-card-title あなたの参加予定日
        v-card-text
          template(v-if="$apollo.queries.fetchData.loading")
            span ロード中...
          template(v-else)
            v-chip(
              v-for="day in fetchData.me.entries",
              :key="day",
              color="primary",
              text-color="black",
              small
            ) {{ day }}日目
    v-flex
      v-card
        v-card-title リクエスト受付締め切り日時
        v-card-text
          template(v-if="$apollo.queries.fetchData.loading")
            span ロード中...
          template(v-else)
            span(v-for="deadline in deadlines", :key="deadline.day")
              template(v-if="deadline.day === 0")
                span 企業: {{ formatDatetime(deadline.datetime) }}
                br
              template(v-else)
                span {{ deadline.day }}日目: {{ formatDatetime(deadline.datetime) }}
                br
</template>

<script>
import gql from "graphql-tag";
import dayjs from "dayjs";

const getMe = gql`
  query {
    me {
      entries
    }
    deadlines {
      day
      datetime
      over
    }
  }
`;

export default {
  name: "home",
  data: function () {
    return {
      fetchData: {
        me: {
          entries: [],
        },
        deadlines: [],
      },
    };
  },
  computed: {
    deadlines: function () {
      return this.fetchData.deadlines;
    },
  },
  apollo: {
    fetchData: {
      query: getMe,
      update: (data) => data,
      fetchPolicy: "network-only",
    },
  },
  methods: {
    formatDatetime: (dt) => dayjs(dt).format("MM/DD HH:mm"),
  },
};
</script>
