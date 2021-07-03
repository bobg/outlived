import React, { useCallback, useState } from 'react'

import {
  Button,
  FormControlLabel,
  Paper,
  Switch,
  Tooltip,
} from '@material-ui/core'

import { LoggedInUser } from './LoggedInUser'
import { LoggedOutUser } from './LoggedOutUser'

import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'

interface Props {
  user: UserData | null
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

export const User = (props: Props) => {
  const { user, setUser, setAlert } = props

  return user ? (
    <LoggedInUser user={user} setUser={setUser} setAlert={setAlert} />
  ) : (
    <LoggedOutUser setUser={setUser} setAlert={setAlert} />
  )
}
