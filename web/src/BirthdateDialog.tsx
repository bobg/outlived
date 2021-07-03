import React, { useCallback, useState } from 'react'

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from '@material-ui/core'

interface Props {
  open: boolean
  close: () => void
  onSubmit: (d: Date) => void
}

// TODO: consider using the datepicker from https://material-ui-pickers.dev/

export const BirthdateDialog = (props: Props) => {
  const { open, close, onSubmit } = props

  const [birthdate, setBirthdate] = useState<Date | null>(null)

  return (
    <Dialog open={open}>
      <DialogTitle>Birth date</DialogTitle>
      <form
        onSubmit={(ev: React.FormEvent<HTMLFormElement>) => {
          ev.preventDefault()
          if (birthdate) {
            close()
            onSubmit(birthdate)
          }
        }}
      >
        <DialogContent>
          <TextField
            id='birthdate'
            label='Birth date'
            type='date'
            defaultValue='1988-11-14'
            onChange={(ev: React.ChangeEvent<HTMLInputElement>) => {
              try {
                const d = new Date(ev.target.value)
                if (birthdateValid(d)) {
                  setBirthdate(d)
                }
              } catch (error) {
                // xxx ignore
              }
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={close}>Cancel</Button>
          <Button disabled={!birthdate} type='submit'>
            Submit
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  )
}

const yearMillis =
  365 /* days */ * 24 /* hours */ * 60 /* mins */ * 60 /* secs */ * 1000

const birthdateValid = (d: Date) => {
  const now = new Date()
  const diff = now.valueOf() - d.valueOf()
  if (diff < 13 * yearMillis) {
    // No one under 13, thanks to COPPA.
    return false
  }
  if (diff > 150 * yearMillis) {
    // No one over 150, the maximum human lifespan!
    // (https://www.scientificamerican.com/article/humans-could-live-up-to-150-years-new-research-suggests/)
    return false
  }
  return true
}
