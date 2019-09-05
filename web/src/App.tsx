import React from 'react';
import logo from './logo.svg';
import './App.css';

interface Props {
}

interface State {
  data: any // xxx
}

class App extends React.Component{
  constructor(props) {
    super(props)
    this.getData()
  }

  private getData = async () => {
    // xxx server request
    this.setState({data})
  }

  public render() {
    const { data } = this.state

    return (
        <div className="App">
        <header className="App-header">
        <img src={logo} className="App-logo" alt="Outlived" />
        </header>
        {data.user && (
            <SignedIn user={data.user} />
        )}
        <Figures figures={data.figures} user={data.user} />
      </div>
    );
  }
}

export default App;
