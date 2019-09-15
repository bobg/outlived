import React from 'react';

interface Props {
  figures: any[]
  user: any
}

export class Figures extends React.Component<Props> {
  public render = () => (
    <div>
      {this.props.figures.map((fig: any) => (
        <div>
          <a className="figure" target="_blank" rel="noopener noreferrer" href={fig.href}>
            {fig.imgSrc && (
              <img className="img64" src={fig.imgSrc} alt={fig.imgAlt} />
            )}
            {fig.name}<br />
            {fig.desc}{fig.desc && (
              <br />
            )}
            {fig.born}&mdash;{fig.died}
          </a>
        </div>
      ))}
    </div>
  )
}
