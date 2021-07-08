import React, { useCallback, useState } from 'react'

import { Paper } from '@material-ui/core'
import { makeStyles, Theme } from '@material-ui/core/styles'

import { LoggedInUser } from './LoggedInUser'
import { LoggedOutUser } from './LoggedOutUser'

import { post } from './post'
import { UserData } from './types'

interface Props {
  user: UserData | null
  setUser: (user: UserData) => void
  setAlert: (alert: string, severity?: 'error' | 'info') => void
}

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    background: theme.palette.primary.dark,
    color: theme.palette.primary.contrastText,
    padding: '0.25rem',
  },
}))

export const User = (props: Props) => {
  const { user, setUser, setAlert } = props

  const classes = useStyles()

  return (
    <Paper className={classes.paper} elevation={0}>
      {user ? (
        <LoggedInUser user={user} setUser={setUser} setAlert={setAlert} />
      ) : (
        <LoggedOutUser setUser={setUser} setAlert={setAlert} />
      )}
    </Paper>
  )
}
