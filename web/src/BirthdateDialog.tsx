import React, { useCallback, useState } from 'react'

import {
  Dialog,
  DialogActions,
  DialogContent,
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
    <Dialog>
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
              // xxx parse ev.target.value as a date
              setBirthdate(d)
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
