import { UI_CONSTANTS } from '../../utils/constants';

interface HeaderProps {
  currentTime: string;
  currentDate: string;
}

export function Header({ currentTime, currentDate }: HeaderProps) {
  return (
    <div className="header">
      <div className="header-left">
        <span className="prompt-symbol">&#62;</span>
        <span className="title">NEURAL.INTERFACE.{UI_CONSTANTS.INTERFACE_VERSION}</span>
      </div>
      <div className="header-right">
        <span className="time">{currentTime}</span>
        <br />
        <span className="date">{currentDate}</span>
      </div>
    </div>
  );
}