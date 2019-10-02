import React from 'react'
import './App.css'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'react-toggle/style.css'

import { Figures } from './Figures'
import { post } from './post'
import { LoggedInUser, LoggedOutUser } from './User'
import { Data, FigureData, UserData } from './types'
import { tzname } from './tz'

interface State {
  figures: FigureData[]
  today?: string
  user?: UserData
}

class App extends React.Component<{}, State> {
  public state: State = { figures: [] }

  private getData = async () => {
    const resp = await post('/s/data', {
      tzname: tzname(),
    })
    const data = (await resp.json()) as Data
    const { figures, today, user } = data
    this.setState({ figures, today, user })
  }

  public componentDidMount = () => this.getData()

  private onLogin = (user: UserData) => this.setState({ user })

  public render() {
    const { figures, today, user } = this.state

    return (
      <div className='App'>
        <header>Outlived</header>
        {user && <LoggedInUser user={user} />}
        {!user && <LoggedOutUser onLogin={this.onLogin} />}
        {today && <div id='today'>Today is {today}.</div>}
        {figures && <Figures figures={figures} user={user} />}
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
      </div>
    )
  }
}

export default App
