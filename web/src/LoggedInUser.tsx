import React, { useCallback, useState } from 'react'

import {
  Button,
  FormControlLabel,
  Paper,
  Switch,
  Tooltip,
  Typography,
} from '@material-ui/core'
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles'

import { post } from './post'
import { UserData } from './types'

interface Props {
  user: UserData
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

const useStyles = (theme: Theme) => makeStyles({
  paper: {
    background: theme.palette.primary.main,
    color: theme.palette.primary.contrastText,
  },
})

export const LoggedInUser = (props: Props) => {
  const { user, setUser, setAlert } = props

  const { active, csrf, email, verified } = user
  const [receivingMail, setReceivingMail] = useState(verified && active)
  const [reverified, setReverified] = useState(false)

  const theme = useTheme()
  const classes = useStyles(theme)()

  const doSetReceivingMail = async (checked: boolean) => {
    try {
      await post('/s/setactive', { csrf, active: checked })
      setReceivingMail(checked)
    } catch (error) {
      setAlert(error.message)
    }
  }

  const reverify = async () => {
    try {
      await post('/s/reverify', { csrf })
      setReverified(true)
    } catch (error) {
      setAlert(error.message)
    }
  }

  return (
    <Paper className={classes.paper}>
      <div>
        <form method='POST' action='/s/logout'>
          <Typography variant='caption'>Logged in as {email}.</Typography>
          <input type='hidden' name='csrf' value={csrf} />
          <Button type='submit' variant='outlined' color='secondary'>
            Log out
          </Button>
        </form>
      </div>
      <Tooltip title='Up to one message per day showing the notable figures you’ve just outlived.'>
        <FormControlLabel
          label='Receive Outlived mail?'
          control={
            <Switch
              checked={receivingMail}
              disabled={!verified}
              onChange={(
                event: React.ChangeEvent<HTMLInputElement>,
                checked: boolean
              ) => doSetReceivingMail(checked)}
            />
          }
        />
      </Tooltip>
      {!verified &&
        (reverified ? (
          <div id='reverified'>
            Check your e-mail for a verification message from Outlived.
          </div>
        ) : (
          <div id='unconfirmed'>
            You have not yet confirmed your e-mail address.
            <br />
            <Button id='resend-button' onClick={reverify}>
              Resend verification
            </Button>
          </div>
        ))}
    </Paper>
  )
}
