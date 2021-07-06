import React, { useCallback, useState } from 'react'

import { Paper } from '@material-ui/core'
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles'

import { LoggedInUser } from './LoggedInUser'
import { LoggedOutUser } from './LoggedOutUser'

import { post } from './post'
import { UserData } from './types'

interface Props {
  user: UserData | null
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

const useStyles = (theme: Theme) =>
  makeStyles({
    paper: {
      background: theme.palette.primary.dark,
      color: theme.palette.primary.contrastText,
      padding: '0.25rem',
    },
  })

export const User = (props: Props) => {
  const { user, setUser, setAlert } = props

  const theme = useTheme()
  const classes = useStyles(theme)()

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
