import React from "react";

interface Props {
  csrf: string;
}

export class LogoutButton extends React.Component<Props> {
  public render = () => (
    <form method="POST" action="/logout">
      <input type="hidden" value={this.props.csrf} />
      <button type="submit">Log out</button>
    </form>
  );
}
