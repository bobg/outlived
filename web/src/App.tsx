import React, { useCallback, useState } from 'react'
import ReactDOM from 'react-dom'

import {
  AppBar,
  Box,
  CircularProgress,
  Link,
  Snackbar,
  ThemeProvider,
  Typography,
} from '@material-ui/core'
import { Alert } from '@material-ui/lab'
import { createMuiTheme } from '@material-ui/core/styles'

import { Figures } from './Figures'
import { TopBar } from './TopBar'
import { User } from './User'

import { post } from './post'
import { Data, FigureData, UserData } from './types'
import { tzname } from './tz'

export const App = () => {
  const [alert, setAlert] = useState('')
  const [figures, setFigures] = useState<FigureData[]>([])
  const [loaded, setLoaded] = useState(false)
  const [today, setToday] = useState('')
  const [user, setUser] = useState<UserData | null>(null)

  const theme = createMuiTheme({
    typography: {
      button: {
        textTransform: 'none',
      },
    },
  })

  const getData = async () => {
    try {
      const resp = await post('/s/data', { tzname: tzname() })
      const data = (await resp.json()) as Data
      if (!data.figures) {
        setAlert('Error: received no figures from server')
      } else {
        setFigures(data.figures)
        setToday(data.today)
        setUser(data.user)
        setLoaded(true)
      }
    } catch (error) {
      setAlert(`Error loading data: ${error.message}`)
    }
  }

  if (!loaded && !alert) {
    getData()
  }

  return (
    <ThemeProvider theme={theme}>
      <TopBar user={user} setUser={setUser} setAlert={setAlert} />
      <Snackbar open={!!alert} onClose={() => setAlert('')}>
        <Alert severity='error'>{alert}</Alert>
      </Snackbar>
      {loaded ? (
        <>
          <Figures figures={figures} today={today} user={user} />
          <Box alignItems='center' justifyContent='center' textAlign='center'>
            <Typography paragraph={true} variant='caption'>
              Data supplied by{' '}
              <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://en.wikipedia.org/'
              >
                Wikipedia
              </Link>
              , the free encyclopedia.
            </Typography>
            <Typography paragraph={true} variant='caption'>
              Some graphic design elements supplied by Suzanne Glickstein.
              Thanks Suze!
            </Typography>
            <Typography paragraph={true} variant='caption'>
              Curious about how this site works? Read the source at{' '}
              <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://github.com/bobg/outlived/'
              >
                github.com/bobg/outlived
              </Link>
              !
            </Typography>
          </Box>
        </>
      ) : (
        <CircularProgress />
      )}
    </ThemeProvider>
  )
}

ReactDOM.render(<App />, document.getElementById('root'))
