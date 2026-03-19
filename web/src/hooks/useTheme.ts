import { useState, useEffect } from 'react';

export type Theme = 'dark' | 'light';

function migrateTheme(stored: string | null): Theme {
  if (stored === 'green') return 'dark';
  if (stored === 'blue') return 'light';
  if (stored === 'dark' || stored === 'light') return stored;
  return 'dark';
}

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(() => {
    return migrateTheme(localStorage.getItem('theme'));
  });

  useEffect(() => {
    const root = document.documentElement;
    if (theme === 'dark') {
      root.removeAttribute('data-theme');
    } else {
      root.setAttribute('data-theme', theme);
    }
    localStorage.setItem('theme', theme);
  }, [theme]);

  return { theme, setTheme };
}
