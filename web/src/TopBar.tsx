import React, { useCallback, useState } from 'react'

import { AppBar, Toolbar } from '@material-ui/core'

import { User } from './User'
import { UserData } from './types'

interface Props {
  user: UserData | null
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

export const TopBar = (props: Props) => {
  const { user, setUser, setAlert } = props

  return (
    <AppBar position='static'>
      <Toolbar>
        <img src='outlived.png' alt='Outlived' width='80%' />
        <User user={user} setUser={setUser} setAlert={setAlert} />
      </Toolbar>
    </AppBar>
  )
}
