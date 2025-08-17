import { useState, useEffect } from 'react';
import { UI_CONSTANTS } from '../utils/constants';

export function useTime() {
  const [currentTime, setCurrentTime] = useState('');
  const [currentDate, setCurrentDate] = useState('');

  useEffect(() => {
    const updateTime = () => {
      const now = new Date();
      
      setCurrentTime(now.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: true
      }));

      setCurrentDate(now.toLocaleDateString('en-US', { 
        month: '2-digit', 
        day: '2-digit', 
        year: 'numeric' 
      }));
    };

    // Update immediately
    updateTime();

    // Update every second
    const interval = setInterval(updateTime, UI_CONSTANTS.UPDATE_INTERVAL);

    return () => clearInterval(interval);
  }, []);

  return { currentTime, currentDate };
}