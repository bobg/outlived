import React, { useCallback, useState } from 'react'
import ReactDOM from 'react-dom'

import {
  AppBar,
  Box,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  Link,
  Snackbar,
  ThemeProvider,
  Typography,
} from '@material-ui/core'
import { Alert } from '@material-ui/lab'
import {
  createTheme,
  makeStyles,
  Theme,
  useTheme,
} from '@material-ui/core/styles'

import { Figures } from './Figures'
import { TopBar } from './TopBar'
import { User } from './User'

import { post } from './post'
import { Data, FigureData, UserData } from './types'
import { tzname } from './tz'

// https://paletton.com/#uid=22K0u0kfvtS5WOhaDChkaoXohkz
const theme = createTheme({
  palette: {
    primary: {
      light: '#A5E095',
      main: '#7BC867',
      dark: '#54A73E',
      contrastText: '#378A21',
    },
    secondary: {
      light: '#F7A5AF',
      main: '#E47683',
      dark: '#BF4755',
      contrastText: '#9D2634',
    },
    background: {
      default: '#CCEFC3',
      paper: '#CCEFC3',
    },
  },
  typography: {
    button: {
      textTransform: 'none',
    },
  },
})

const useStyles = (theme: Theme) =>
  makeStyles({
    today: {
      backgroundColor: theme.palette.primary.light,
      borderWidth: '4px',
      color: theme.palette.primary.dark,
      fontSize: '1.2rem',
      margin: '1rem',
      textAlign: 'center',
      width: 'fit-content',
    },
  })

export const App = () => {
  const [alert, setAlert] = useState('')
  const [figures, setFigures] = useState<FigureData[]>([])
  const [loaded, setLoaded] = useState(false)
  const [today, setToday] = useState('')
  const [user, setUser] = useState<UserData | null>(null)

  const classes = useStyles(theme)()

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
          <Box display='flex' justifyContent='center' m='auto'>
            <Card className={classes.today} raised={true} variant='outlined'>
              <CardContent>
                <CardHeader title={`Today is ${today}`} />
                {user ? (
                  <Typography>
                    You were born on {user.born}, which was{' '}
                    {user.daysAlive.toLocaleString()} days ago
                    <br />({user.yearsDaysAlive}).
                  </Typography>
                ) : (
                  undefined
                )}
              </CardContent>
            </Card>
          </Box>
          <Figures
            diedToday={figures}
            outlived={user ? user.figures : undefined}
          />
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
        <Box display='flex' justifyContent='center' m='2em'>
          <CircularProgress />
        </Box>
      )}
    </ThemeProvider>
  )
}

ReactDOM.render(<App />, document.getElementById('root'))
