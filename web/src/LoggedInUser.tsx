import React, { useCallback, useState } from 'react'

import {
  Box,
  Button,
  Dialog,
  DialogTitle,
  FormControlLabel,
  Link,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Paper,
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
  setAlert: (alert: string, severity?: 'error'|'info') => void
}

const useStyles = (theme: Theme) =>
  makeStyles({
    logout: {
      padding: '.25rem',
    },
    email: {
      color: theme.palette.primary.contrastText,
      cursor: 'pointer',
    },
    paper: {
      margin: '2px',
      padding: '2px',
    },
    receiveLabel: {
      fontSize: theme.typography.caption.fontSize,
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
          <Paper className={classes.paper} elevation={0}>
            <Box display='flex'>
              <Box textAlign='center'>
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
                  color='primary'
                  size='small'
                >
                  Log&nbsp;out
                </Button>
              </Box>
            </Box>
          </Paper>
        </form>
      </div>

      <Paper className={classes.paper} elevation={0}>
        <Box textAlign='center'>
          <Tooltip title='Up to one message per day showing the notable figures you’ve just outlived.'>
            <Box>
              <FormControlLabel
                classes={{label: classes.receiveLabel}}
                control={
                  <Switch
                    size='small'
                    checked={receivingMail}
                    onChange={(
                      event: React.ChangeEvent<HTMLInputElement>,
                      checked: boolean
                    ) => doSetReceivingMail(checked)}
                  />
                }
                disabled={!verified}
                label='Receive Outlived mail?'
              />
            </Box>
          </Tooltip>
          {!verified ? (
            <Typography variant='caption'>
              {reverified ? (
                <div id='reverified'>
                  Check your e-mail for a verification message from Outlived.
                </div>
              ) : (
                <>
                  You have not yet confirmed your e-mail address.
                  <br />
                  <Button
                    id='resend-button'
                    onClick={reverify}
                    size='small'
                    variant='outlined'
                    color='primary'
                  >
                    Resend&nbsp;verification
                  </Button>
                </>
              )}
            </Typography>
          ) : (
            undefined
          )}
        </Box>
      </Paper>

      <Dialog open={settingsOpen} onClose={() => setSettingsOpen(false)}>
        <DialogTitle>Settings</DialogTitle>
        <List>
          <ListItem button onClick={chooseBirthdate}>
            <ListItemIcon>
              <Event />
            </ListItemIcon>
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
    </>
  )
}
