import React, { useCallback, useState } from 'react'

import { Button, Link, Switch, Tooltip, Typography } from '@material-ui/core'
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles'

import { post } from './post'
import { UserData } from './types'

interface Props {
  user: UserData
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

const useStyles = (theme: Theme) =>
  makeStyles({
    logout: {
      background: theme.palette.secondary.light,
      color: theme.palette.secondary.contrastText,
      fontSize: theme.typography.caption.fontSize,
      padding: '.25rem',
    },
    email: {
      color: theme.palette.primary.contrastText,
      cursor: 'pointer',
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

  const onEmailClick = () => {
    console.log(email)
  }

  return (
    <>
      <div>
        <form method='POST' action='/s/logout'>
          <input type='hidden' name='csrf' value={csrf} />
          <Typography variant='caption'>
            Logged in as <Link className={classes.email} onClick={onEmailClick}>{email}</Link>.{' '}
            <Button
              className={classes.logout}
              type='submit'
              variant='outlined'
              color='secondary'
            >
              Log out
            </Button>
          </Typography>
        </form>
      </div>
      <Tooltip title='Up to one message per day showing the notable figures you’ve just outlived.'>
        <Typography variant='caption'>
          Receive Outlived mail?{' '}
          <Switch
            size='small'
            checked={receivingMail}
            disabled={!verified}
            onChange={(
              event: React.ChangeEvent<HTMLInputElement>,
              checked: boolean
            ) => doSetReceivingMail(checked)}
          />
        </Typography>
      </Tooltip>
      {!verified ? (
        reverified ? (
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
        )
      ) : null}
    </>
  )
}
