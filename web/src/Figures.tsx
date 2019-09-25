import React from 'react'
import { Tab, Tabs } from 'react-bootstrap'

import { FigureData, UserData } from './types'

interface Props {
  figures: FigureData[]
  user?: UserData
}

interface State {
  activeTab: string
}

export class Figures extends React.Component<Props, State> {
  public state: State = {activeTab: '2'}

  private handleTab = (activeTab: string) => {
    this.setState({ activeTab })
  }

  public render = () => {
    const { figures, user } = this.props
    if (!figures) {
      return null
    }

    if (user) {
      return (
        <Tabs
          id="figures"
          activeKey={this.state.activeTab}
          onSelect={this.handleTab}
        >
          <Tab eventKey={1} title="Died on this date">
            {renderFigs(figures, true)}
          </Tab>
          <Tab eventKey={2} title="You have recently outlived">
            {renderFigs(user.figures, false)}
          </Tab>
        </Tabs>
      )
    }

    return (
      <div>
        <p>Died on this date:</p>
        {renderFigs(figures, true)}
      </div>
    )
  }
}

const renderFigs = (figs: FigureData[], showAge: boolean) => (
  <ul className="grid">
    {figs.map((fig: FigureData) => (
      <li>
        <a
          className="figure"
          target="_blank"
          rel="noopener noreferrer"
          href={fig.href}
        >
          {fig.imgSrc && (
            <span>
              <img className="img128" src={fig.imgSrc} alt={fig.imgAlt} />
              <br />
            </span>
          )}
          {fig.name}
        </a>
        <br />
        {fig.desc}
        {fig.desc && <br />}
        {fig.born}&mdash;{fig.died}
        {showAge && <br />}
        {showAge && '(' + fig.yearsDaysAlive + ')'}
      </li>
    ))}
  </ul>
)