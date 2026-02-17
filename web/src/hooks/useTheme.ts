import { useState, useEffect } from 'react';

export type Theme = 'green' | 'blue';

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(() => {
    return (localStorage.getItem('theme') as Theme) || 'green';
  });

  useEffect(() => {
    const root = document.documentElement;
    if (theme === 'green') {
      root.removeAttribute('data-theme');
    } else {
      root.setAttribute('data-theme', theme);
    }
    localStorage.setItem('theme', theme);
  }, [theme]);

  return { theme, setTheme };
}
