import React, { useCallback, useState } from 'react'

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
} from '@material-ui/core'

interface Props {
  close: () => void
  mode: string
  open: boolean
  onSubmit: (pw: string) => void
  onForgot?: () => void
}

export const Password = (props: Props) => {
  const [pw, setPW] = useState('')

  const { close, onSubmit, onForgot, mode, open } = props

  const doForgot = () => {
    if (!onForgot) {
      return
    }
    onForgot()
    close()
  }

  return (
    <Dialog open={open}>
      <DialogTitle>Password</DialogTitle>
      <form
        onSubmit={(ev: React.FormEvent<HTMLFormElement>) => {
          ev.preventDefault()
          if (pw !== '') {
            close()
            onSubmit(pw)
          }
        }}
      >
        <DialogContent>
          <DialogContentText>
            {mode === 'login' ? 'Enter your password' : 'Choose a password'}
          </DialogContentText>
          <TextField
            defaultValue=''
            autoFocus
            margin='dense'
            id='password'
            type='password'
            label='Password'
            fullWidth
            onChange={(ev: React.ChangeEvent<HTMLInputElement>) => {
              setPW(ev.target.value)
            }}
          >
            {pw}
          </TextField>
        </DialogContent>
        <DialogActions>
          <Button onClick={close}>Cancel</Button>
          {onForgot ? (
            <Button onClick={doForgot}>Forgot password</Button>
          ) : null}
          <Button disabled={pw === ''} type='submit'>
            Submit
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  )
}
