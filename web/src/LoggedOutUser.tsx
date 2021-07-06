import React, { useCallback, useState } from 'react'

import {
  Box,
  Button,
  ButtonGroup,
  Paper,
  TextField,
  Typography,
} from '@material-ui/core'
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles'

import { BirthdateDialog } from './BirthdateDialog'
import { Password } from './Password'

import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'
import { datestr } from './util'

interface Props {
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

const useStyles = (theme: Theme) =>
  makeStyles({
    paper: {
      background: theme.palette.primary.main,
      color: theme.palette.primary.contrastText,
    },
  })

export const LoggedOutUser = (props: Props) => {
  const { setUser, setAlert } = props

  const [birthDate, setBirthDate] = useState<Date | null>(null)
  const [birthdateDialogOpen, setBirthdateDialogOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [pwOpen, setPWOpen] = useState(false)
  const [pwMode, setPWMode] = useState('')

  const theme = useTheme()
  const classes = useStyles(theme)()

  const onLoginButton = () => {
    setPWOpen(true)
    setPWMode('login')
  }

  const onSignupButton = () => {
    setBirthdateDialogOpen(true)
    setPWMode('signup')
  }

  const onSubmitBirthdate = (d: Date) => {
    setBirthDate(d)
    setBirthdateDialogOpen(false)
    setPWOpen(true)
  }

  const onSubmitPW = (pw: string) => {
    if (pwMode === 'login') {
      doLogin(pw)
    } else if (pwMode === 'signup') {
      doSignup(pw)
    }
  }

  const doLogin = async (pw: string) => {
    if (!email || !pw) {
      return
    }
    try {
      const resp = await post('/s/login', {
        email,
        forgot: false,
        password: pw,
        tzname: tzname(),
      })
      const user = (await resp.json()) as UserData
      setUser(user)
    } catch (error) {
      setAlert(`Login failed: ${error.message}`)
    }
  }

  const doSignup = async (pw: string) => {
    if (!email || !birthDate || !pw) {
      return
    }
    try {
      const resp = await post('/s/signup', {
        email,
        password: pw,
        born: datestr(birthDate),
        tzname: tzname(),
      })
      const user = (await resp.json()) as UserData
      setUser(user)
    } catch (error) {
      setAlert(`Signup failed: ${error.message}`)
    }
  }

  const doForgot = async () => {
    if (!email) {
      return
    }
    try {
      const resp = await post('/s/login', {
        email,
        forgot: true,
        tzname: tzname(),
      })
      setAlert('Check your e-mail for a password-reset message.') // xxx non-error alert
    } catch (error) {
      setAlert(`Login failed: ${error.message}`)
    }
  }

  return (
    <>
      <Paper className={classes.paper}>
        <Typography variant='caption'>
          Log in to see whom you’ve recently outlived.
        </Typography>
        <TextField
          autoFocus
          defaultValue=''
          onChange={(ev: React.ChangeEvent<HTMLInputElement>) => {
            setEmail(ev.target.value)
          }}
          placeholder='E-mail address'
        />
        <Box
          display='flex'
          flexDirection='column'
          alignItems='center'
          justifyContent='center'
        >
          <ButtonGroup>
            <Button
              disabled={!emailValid(email)}
              onClick={onLoginButton}
              color='primary'
            >
              Log in
            </Button>
            <Button
              disabled={!emailValid(email)}
              onClick={onSignupButton}
              color='primary'
            >
              Sign up
            </Button>
          </ButtonGroup>
        </Box>
      </Paper>
      <BirthdateDialog
        open={birthdateDialogOpen}
        close={() => setBirthdateDialogOpen(false)}
        onSubmit={onSubmitBirthdate}
        defaultVal={'1988-11-14'}
      />
      <Password
        open={pwOpen}
        close={() => setPWOpen(false)}
        mode={pwMode}
        onSubmit={onSubmitPW}
        onForgot={doForgot}
      />
    </>
  )
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp?: string) => {
  return inp && /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp)
}
