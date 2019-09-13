import React from 'react';
import logo from './logo.svg';
import './App.css';

interface State {
  data: any // xxx
}

class App extends React.Component<State> {
  constructor(props) {
    super(props)
    this.state = {data: {}}
  }

  private getData = async () => {
    const req = new XMLHttpRequest()
    return new Promise((resolve, reject) => {
      req.open('POST', xxxurl, true) // xxx true?
      req.send(xxx)
    })
  }


    // xxx server request
    this.setState({data})
  }

  public componentDidMount = () => {
    this.getData()
  }

  public render() {
    const { data } = this.state

    return (
      <div className="App">
        <header className="App-header">
          <img src={logo} className="App-logo" alt="Outlived" />
        </header>
        <User user={data.user} />
        <Figures figures={data.figures} user={data.user} />
      </div>
    );
  }
}

export default App;
