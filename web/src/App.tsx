import React from 'react'
import './App.css'
import 'bootstrap/dist/css/bootstrap.min.css'

import { Figures } from './Figures'
import { post } from './post'
import { User } from './User'
import { Data, UserData } from './types'
import { tzname } from './tz'

interface State {
  data: Data
}

class App extends React.Component<{}, State> {
  constructor(props: any) {
    super(props)
    this.state = { data: {} }
  }

  private getData = async () => {
    const resp = await post('/s/data', {
      tzname: tzname(),
    })
    const data = (await resp.json()) as Data
    this.setState({ data })
  }

  public componentDidMount = () => this.getData()

  private onLogin = (user: UserData) => {
    const { data } = this.state
    this.setState({ data: { ...data, user } })
  }

  public render() {
    const { figures, today, user } = this.state.data

    return (
      <div className='App'>
        <header>Outlived</header>
        <User onLogin={this.onLogin} user={user} />
        {today && <div>Today is {today}.</div>}
        {figures && <Figures figures={figures} user={user} />}
      </div>
    )
  }
}

export default App
