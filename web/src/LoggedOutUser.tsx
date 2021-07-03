import React, { useCallback, useState } from 'react'

import { Button, Paper, TextField } from '@material-ui/core'

import { BirthdateDialog } from './BirthdateDialog'
import { Password } from './Password'

import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'

interface Props {
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

export const LoggedOutUser = (props: Props) => {
  const { setUser, setAlert } = props

  const [birthDate, setBirthDate] = useState<Date | null>(null)
  const [birthdateDialogOpen, setBirthdateDialogOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [pwOpen, setPWOpen] = useState(false)
  const [pwMode, setPWMode] = useState('')

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

  return (
    <>
      <Paper>
        <TextField
          autoFocus
          defaultValue=''
          onChange={(ev: React.ChangeEvent<HTMLInputElement>) => {
            setEmail(ev.target.value)
          }}
        />
        <Button disabled={!emailValid(email)} onClick={onLoginButton}>
          Log in
        </Button>
        <Button disabled={!emailValid(email)} onClick={onSignupButton}>
          Sign up
        </Button>
      </Paper>
      <BirthdateDialog
        open={birthdateDialogOpen}
        close={() => setBirthdateDialogOpen(false)}
        onSubmit={onSubmitBirthdate}
      />
      <Password
        open={pwOpen}
        close={() => setPWOpen(false)}
        mode={pwMode}
        onSubmit={onSubmitPW}
      />
    </>
  )
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp?: string) => {
  return inp && /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp)
}

const datestr = (d: Date) => {
  const y = d.getFullYear()
  const m = 1 + d.getMonth() // come on javascript
  const day = d.getDate() // come ON, javascript
  return y + '-' + ('0' + m).substr(-2) + '-' + ('0' + day).substr(-2)
}
