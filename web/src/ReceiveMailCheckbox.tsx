import React from "react";

interface Props {
  csrf: string;
  user: any;
}

interface State {
  active: boolean;
}

export class ReceiveMailCheckbox extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props);
    this.state = { active: props.user.active };
  }

  public render = () => (
    <div>
      <label htmlFor="active">Receive Outlived mail?</label>
      <input
        type="checkbox"
        id="active"
        name="active"
        checked={this.state.active}
        disabled={!this.props.user.verified}
      />
    </div>
  );
}
