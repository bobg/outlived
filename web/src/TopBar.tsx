import React, { useCallback, useState } from 'react'

import { AppBar, Box, Toolbar } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'

import { User } from './User'
import { UserData } from './types'

interface Props {
  user: UserData | null
  setUser: (user: UserData) => void
  setAlert: (alert: string, severity?: string) => void
}

const useStyles = makeStyles({
  logo: {
    flexGrow: 1,
  },
  user: {
    // minWidth: '12em',
  },
})

export const TopBar = (props: Props) => {
  const { user, setUser, setAlert } = props

  const classes = useStyles()

  return (
    <AppBar position='static'>
      <Toolbar>
        <Box display='flex'>
          <Box className={classes.logo}>
            <img src='outlived.png' alt='Outlived' width='80%' />
          </Box>
          <Box className={classes.user}>
            <User user={user} setUser={setUser} setAlert={setAlert} />
          </Box>
        </Box>
      </Toolbar>
    </AppBar>
  )
}
