import React, { useCallback, useState } from 'react'

import {
  Box,
  Button,
  Dialog,
  DialogTitle,
  Link,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Switch,
  Tooltip,
  Typography,
} from '@material-ui/core'
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles'
import { Event } from '@material-ui/icons'

import { BirthdateDialog } from './BirthdateDialog'

import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'
import { datestr } from './util'

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
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [birthdateOpen, setBirthdateOpen] = useState(false)

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

  const onNewBirthdate = async (newDate: Date) => {
    const newStr = datestr(newDate)
    if (newStr === user.bornyyyymmdd) {
      return
    }
    try {
      const resp = await post('/s/setbirthdate', {
        csrf,
        tzname: tzname(),
        newdate: newStr,
      })
      const user = (await resp.json()) as UserData
      setUser(user)
    } catch (error) {
      setAlert(error.message)
    }
  }

  const chooseBirthdate = () => {
    setSettingsOpen(false)
    setBirthdateOpen(true)
  }

  return (
    <>
      <div>
        <form method='POST' action='/s/logout'>
          <Box display='flex'>
            <Box>
              <input type='hidden' name='csrf' value={csrf} />
              <Typography variant='caption'>
                Logged in as{' '}
                <Tooltip title='Tap for settings'>
                  <Link
                    className={classes.email}
                    onClick={() => setSettingsOpen(true)}
                  >
                    {email}
                  </Link>
                </Tooltip>
                .
              </Typography>
            </Box>
            <Box>
              <Button
                className={classes.logout}
                type='submit'
                variant='outlined'
                color='secondary'
              >
                Log&nbsp;out
              </Button>
            </Box>
          </Box>
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
      <Dialog open={settingsOpen} onClose={() => setSettingsOpen(false)}>
        <DialogTitle>Settings</DialogTitle>
        <List>
          <ListItem button onClick={chooseBirthdate}>
            <ListItemText primary='Birth date' />
          </ListItem>
        </List>
      </Dialog>
      <BirthdateDialog
        open={birthdateOpen}
        close={() => setBirthdateOpen(false)}
        onSubmit={onNewBirthdate}
        defaultVal={user.bornyyyymmdd}
      />
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
      ) : (
        undefined
      )}
    </>
  )
}
