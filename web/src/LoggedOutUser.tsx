import React, { useCallback, useState } from 'react'

import { Button, Paper, TextField } from '@material-ui/core'

import { BirthdateDialog } from './BirthdateDialog'
import { Password } from './Password'

import { UserData } from './types'

interface Props {
  setUser: (user: UserData) => void
  setAlert: (alert: string) => void
}

export const LoggedOutUser = (props: Props) => {
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
      <BirthdateDialog open={birthdateDialogOpen} close={() => setBirthdateDialogOpen(false)} onSubmit={onSubmitBirthdate}/>
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
