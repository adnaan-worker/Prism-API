import { theme, ThemeConfig } from 'antd';

export const getAdminTheme = (): ThemeConfig => {
    return {
        algorithm: theme.darkAlgorithm,
        token: {
            colorPrimary: '#0ea5e9',
            colorBgBase: '#000000',
            colorBgContainer: '#111111',
            colorBgElevated: '#1a1a1a',
            colorBorder: '#262626',
            colorText: '#ffffff',
            colorTextSecondary: '#a1a1aa',
            borderRadius: 6, // Slightly sharper for professional look
            fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
        },
        components: {
            Layout: {
                bodyBg: '#050505', // Slightly lighter than pure black to differentiate from header
                headerBg: '#111111',
                siderBg: '#0a0a0a',
            },
            Table: {
                colorBgContainer: '#111111',
                headerBg: '#1a1a1a',
                headerColor: '#a1a1aa',
            },
            Card: {
                colorBgContainer: '#111111',
                colorBorderSecondary: '#262626',
            },
            Menu: {
                itemBg: 'transparent',
                darkItemBg: 'transparent',
                popupBg: '#1a1a1a',
            }
        },
    };
};
