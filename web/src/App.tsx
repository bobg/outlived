import React from 'react'
import './App.css'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'react-toggle/style.css'
import 'semantic-ui-css/semantic.min.css'
import { Loader } from 'semantic-ui-react'

import { Alert, doAlert, setAlertRef } from './Alert'
import { Figures } from './Figures'
import { post } from './post'
import { LoggedInUser, LoggedOutUser } from './User'
import { Data, FigureData, UserData } from './types'
import { tzname } from './tz'

interface State {
  figures: FigureData[]
  loaded: boolean
  today?: string
  user?: UserData
}

class App extends React.Component<{}, State> {
  public state: State = { figures: [], loaded: false }

  private getData = async () => {
    try {
      const resp = await post('/s/data', {
        tzname: tzname(),
      })
      const data = (await resp.json()) as Data
      const { figures, today, user } = data
      this.setState({ figures, loaded: true, today, user })
    } catch (error) {
      doAlert('Error loading data. Please try reloading this page in a moment.')
    }
  }

  public componentDidMount = () => this.getData()

  private onLogin = (user: UserData) => this.setState({ user })

  public render() {
    const { figures, loaded, today, user } = this.state

    return (
      <div className='App'>
        <Alert ref={(r: Alert) => setAlertRef(r)} />
        <header>
          <img src='outlived.png' alt='Outlived' width="80%" />
        </header>
        {loaded ? (
          <>
            {user && <LoggedInUser user={user} />}
            {!user && <LoggedOutUser onLogin={this.onLogin} />}
            {figures.length > 0 && (
              <Figures figures={figures} today={today} user={user} />
            )}
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
              Some graphic design elements supplied by Suzanne Glickstein. Thanks Suze!
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
          </>
        ) : (
          <Loader active size='large' />
        )}
      </div>
    )
  }
}

export default App
