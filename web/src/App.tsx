import React from 'react'
import logo from './logo.svg'
import './App.css'

import { Figures } from './Figures'
import { User } from './User'

interface State {
  data: any
}

class App extends React.Component<{}, State> {
  constructor(props: any) {
    super(props)
    this.state = { data: {} }
  }

  private getData = async () => {
    const setState = this.setState

    fetch('/s/data', {
      method: 'POST',
      credentials: 'same-origin',
    }).then((data: any) => {
      setState({ data })
    })
  }

  public componentDidMount = () => this.getData()

  public render() {
    const { data } = this.state

    return (
      <div className='App'>
        <header className='App-header'>
          <img src={logo} className='App-logo' alt='Outlived' />
        </header>
        <User csrf={data.csrf} user={data.user} />
        <Figures figures={data.figures} user={data.user} />
      </div>
    )
  }
}

export default App
