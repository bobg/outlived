import React, { useCallback, useState } from 'react'
import ReactDOM from 'react-dom'

import { AppBar, CircularProgress, Link, Typography } from '@material-ui/core'
import { Alert } from '@material-ui/lab'

import { Figures } from './Figures'
import { post } from './post'
import { User } from './User'
import { Data, FigureData, UserData } from './types'
import { tzname } from './tz'

export const App = () => {
  const [alert, setAlert] = useState('')
  const [figures, setFigures] = useState<FigureData[]>([])
  const [loaded, setLoaded] = useState(false)
  const [today, setToday] = useState('')
  const [user, setUser] = useState<UserData | null>(null)

  const getData = async () => {
    try {
      const resp = await post('/s/data', { tzname: tzname() })
      const data = (await resp.json()) as Data
      setFigures(data.figures)
      setToday(data.today)
      setUser(data.user)
    } catch (error) {
      setAlert(`Error loading data: ${error.message}`)
    }
  }

  if (!loaded) {
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
          <Typography variant='body2'>
            <p>
              Data supplied by{' '}
              <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://en.wikipedia.org/'
              >
                Wikipedia
              </Link>
              , the free encyclopedia.
            </p>
            <p>
              Some graphic design elements supplied by Suzanne Glickstein.
              Thanks Suze!
            </p>
            <p>
              Curious about how this site works? Read the source at{' '}
              <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://github.com/bobg/outlived/'
              >
                github.com/bobg/outlived
              </Link>
              !
            </p>
          </Typography>
        </>
      ) : (
        <CircularProgress />
      )}
    </ThemeProvider>
  )
}

ReactDOM.render(<App />, document.getElementById('root'))

/*

interface State {
  figures: FigureData[]
  loaded: boolean
  showAlert: boolean
  today?: string
  user?: UserData
}

class App extends React.Component<{}, State> {
  public state: State = {
    figures: [],
    loaded: false,
    showAlert: false,
    user: undefined,
  }

  private getData = async () => {
    try {
      const resp = await post('/s/data', {
        tzname: tzname(),
      })
      const data = (await resp.json()) as Data
      const { figures, today, user } = data
      this.setState({ figures, loaded: true, today, user })
    } catch (error) {
      this.setState({ showAlert: true })
    }
  }

  public componentDidMount = () => this.getData()

  private onLogin = (user: UserData) => this.setState({ user })

  public render() {
    const { figures, loaded, showAlert, today, user } = this.state

    return (
      <div className='App'>
        <AppBar position='static'>
          <Toolbar>
            <img src='outlived.png' alt='Outlived' width='80%' />
            <User user={user} onLogin={this.onLogin} />
          </Toolbar>
        </AppBar>
        {showAlert && (
          <Alert severity='error'>
            Error loading data. Please try reloading this page in a moment.
          </Alert>
        )}
        {loaded ? (
          <>
            <Figures figures={figures} today={today} user={user} />
            <Typography variant='body2'>
              <p>
                Data supplied by{' '}
                <a
                  target='_blank'
                  rel='noopener noreferrer'
                  href='https://en.wikipedia.org/'
                >
                  Wikipedia
                </a>
                , the free encyclopedia.
              </p>
              <p>
                Some graphic design elements supplied by Suzanne Glickstein.
                Thanks Suze!
              </p>
              <p>
                Curious about how this site works? Read the source at{' '}
                <a
                  target='_blank'
                  rel='noopener noreferrer'
                  href='https://github.com/bobg/outlived/'
                >
                  github.com/bobg/outlived
                </a>
                !
              </p>
            </Typography>
          </>
        ) : (
          <CircularProgress />
        )}
      </div>
    )
  }
}

export default App

*/
