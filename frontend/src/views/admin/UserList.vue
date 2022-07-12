<template lang="pug">
  v-container(fluid grid-list-md)
    v-card
      v-card-text 右下の+ボタンからメンバー登録ができます。
    v-progress-linear(v-if="$apollo.queries.users.loading" indeterminate)
    template(v-else)
      v-layout(row wrap)
        v-radio-group(v-model="filterDay" row)
          v-radio(label="全員" :value="null")
          v-radio(label="1日目" :value="1")
          v-radio(label="2日目" :value="2")
      v-layout(row wrap)
        template(v-for="user in filteredUsers")
          v-flex(xs12 sm6 md4 lg3)
            v-card
              v-card-title
                span.headline {{ user.displayName }} (@{{ user.name }})
              v-card-text
                v-chip(v-if="user.role === 'PLANNER'" color="green" text-color="white" small) プランナー
                v-chip(v-else-if="user.role === 'ADMIN'" color="red" text-color="white" small) 管理人
                v-chip(v-for="day in user.entries" :key="day" color="primary" text-color="white" small) {{ day }}日目
              v-card-actions
                v-btn(depressed small @click.stop="openEditEntryDialog(user)") 参加日程修正
                v-btn(depressed small @click.stop="openChangePasswordDialog(user)") パスワード変更
                v-btn(depressed small @click.stop="openChangePermissionDialog(user)") 権限変更
    v-btn(fixed dark fab bottom right color="blue darken-2" to="/admin/users/create")
      v-icon add
    v-dialog(v-model="editEntryDialog" width=500 persistent)
      v-card
        v-card-title.headline 参加日程修正
        v-card-text
          p {{ editUser && editUser.displayName }}
          v-layout(row wrap)
            v-checkbox(v-model="editEntries" :value="1" label="1日目")
            v-checkbox(v-model="editEntries" :value="2" label="2日目")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="editEntryDialog = false") キャンセル
          v-btn(color="primary" flat :disabled="sending" :loading="sending" @click.native="editEntry") OK
    v-dialog(v-model="changePasswordDialog" width=500 persistent)
      v-card
        v-card-title.headline パスワード変更
        v-card-text
          p {{ editUser && editUser.displayName }}
          v-form(v-model="valid" ref="passwordForm")
            v-text-field(v-model="newPassword" :append-icon="visiblePassword ? 'visibility_off' : 'visibility'" :type="visiblePassword ? 'text' : 'password'" label="新しいパスワード" :rules="[rules.password]" @click:append="visiblePassword = !visiblePassword")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="changePasswordDialog = false") キャンセル
          v-btn(color="primary" flat :disabled="sending || !valid" :loading="sending" @click.native="changePassword") OK
    v-dialog(v-model="changePermissionDialog" width=500 persistent)
      v-card
        v-card-title.headline 権限変更
        v-card-text
          p {{ editUser && editUser.displayName }}
          v-select(v-model="newRole" :items="roles")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="changePermissionDialog = false") キャンセル
          v-btn(color="primary" flat :disabled="sending" :loading="sending" @click.native="changePermission") OK

</template>

<script>
import gql from 'graphql-tag'

const getDatas = gql`
  query {
    users {
      id
      name
      displayName
      role
      entries
    }
  }
`

const changePassword = gql`
  mutation ($id: Int!, $password: String!) {
    changeUserPassword (userId: $id, password: $password)
  }
`

const changeRole = gql`
  mutation ($id: Int!, $role: Role!) {
    changeUserRole(userId: $id, role: $role) {
      id
      role
    }
  }
`

const changeEntry = gql`
  mutation ($id: Int!, $entries: [Int!]!) {
    changeEntry: changeUserEntries(userId: $id, entries: $entries) {
      id
      entries
    }
  }
`

export default {
  name: 'UserList',
  data: function () {
    return {
      users: [],
      filterDay: null,
      sending: false,
      editEntryDialog: false,
      changePasswordDialog: false,
      changePermissionDialog: false,
      visiblePassword: false,
      valid: false,
      editUser: null,
      editEntries: [],
      newPassword: '',
      newRole: 'USER',
      rules: {
        password: value => /^[a-zA-Z0-9!#$%&()*+,.:;=?@[\]^_{}-]+$/.test(value) || 'パスワードは半角英数文字と記号のみ使えます'
      },
      roles: [
        { text: 'メンバー', value: 'USER' },
        { text: 'プランナー', value: 'PLANNER' },
        { text: '管理人', value: 'ADMIN' }
      ]
    }
  },
  apollo: {
    users: {
      query: getDatas,
      fetchPolicy: 'cache-and-network',
      update: data => data.users
    }
  },
  computed: {
    filteredUsers: function () {
      return this.users.filter(v => this.filterDay ? v.entries.includes(this.filterDay) : true)
    }
  },
  methods: {
    editEntry: async function () {
      this.sending = true
      try {
        await this.$apollo.mutate({
          mutation: changeEntry,
          variables: {
            id: this.editUser.id,
            entries: this.editEntries
          },
          update: (store, { data: { changeEntry } }) => {
            const data = store.readQuery({ query: getDatas })
            data.users.forEach(v => {
              if (v.id === changeEntry.id) {
                v.entries = changeEntry.entries
              }
            })
            store.writeQuery({ query: getDatas, data })
          }
        })
        this.editEntryDialog = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    },
    changePassword: async function () {
      this.sending = true
      try {
        await this.$apollo.mutate({
          mutation: changePassword,
          variables: {
            id: this.editUser.id,
            password: this.newPassword
          }
        })
        this.changePasswordDialog = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    },
    changePermission: async function () {
      this.sending = true
      try {
        this.$apollo.mutate({
          mutation: changeRole,
          variables: {
            id: this.editUser.id,
            role: this.newRole
          },
          update: (store, { data: { changeUserRole } }) => {
            const data = store.readQuery({ query: getDatas })
            data.users.forEach(v => {
              if (v.id === changeUserRole.id) {
                v.role = changeUserRole.role
              }
            })
            store.writeQuery({ query: getDatas, data })
          }
        })
        this.changePermissionDialog = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    },
    openEditEntryDialog (user) {
      this.editUser = user
      this.editEntries = [...user.entries]
      this.editEntryDialog = true
    },
    openChangePasswordDialog (user) {
      this.editUser = user
      this.newPassword = ''
      this.valid = false
      this.$refs.passwordForm.reset()
      this.changePasswordDialog = true
    },
    openChangePermissionDialog (user) {
      this.editUser = user
      this.newRole = user.role
      this.changePermissionDialog = true
    }
  }
}
</script>
