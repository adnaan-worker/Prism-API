import { theme, ThemeConfig } from 'antd';

export const getAdminTheme = (isDarkMode: boolean): ThemeConfig => {
    if (!isDarkMode) {
        return {
            algorithm: theme.defaultAlgorithm,
            token: {
                colorPrimary: '#0ea5e9',
                borderRadius: 8,
                fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
            },
        };
    }

    return {
        algorithm: theme.darkAlgorithm,
        token: {
            colorPrimary: '#0ea5e9', // Sky 500
            colorBgBase: '#000000',   // Black
            colorBgContainer: '#111111', // Card background
            colorBgElevated: '#1a1a1a',  // Floating elements
            colorBorder: '#262626',
            colorText: '#ffffff',
            colorTextSecondary: '#a1a1aa',
            borderRadius: 8,
            fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
        },
        components: {
            Layout: {
                bodyBg: '#000000',
                headerBg: '#111111',
                siderBg: '#0a0a0a',
            },
            Card: {
                colorBgContainer: '#111111',
                colorBorderSecondary: '#262626',
            },
            Button: {
                primaryShadow: '0 2px 0 rgba(14, 165, 233, 0.1)',
            },
            Input: {
                colorBgContainer: '#0a0a0a',
                colorBorder: '#262626',
                activeBorderColor: '#0ea5e9',
            },
            Table: {
                colorBgContainer: '#111111',
                headerBg: '#1a1a1a',
            },
            Modal: {
                contentBg: '#111111',
                headerBg: '#111111',
            },
            Menu: {
                itemBg: 'transparent',
                darkItemBg: 'transparent',
                popupBg: '#1a1a1a',
            }
        },
    };
};
