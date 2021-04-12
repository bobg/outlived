import React from 'react'
import './App.css'

import {
  AppBar,
  CircularProgress,
  Toolbar,
  Typography,
} from '@material-ui/core'
import { Alert } from '@material-ui/lab'

import { Figures } from './Figures'
import { post } from './post'
import { LogInOut } from './User'
import { Data, FigureData, UserData } from './types'
import { tzname } from './tz'

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
            <LogInOut user={user} onLogin={this.onLogin} />
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
