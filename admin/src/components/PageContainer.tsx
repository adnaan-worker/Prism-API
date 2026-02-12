import React from 'react';

interface PageContainerProps {
  /** 页面标题 */
  title: string;
  /** 页面描述 */
  description?: string;
  /** 右侧操作区 */
  extra?: React.ReactNode;
  /** 页面内容 */
  children: React.ReactNode;
}

/**
 * 页面容器 — 统一页面头部结构
 *
 */
const PageContainer: React.FC<PageContainerProps> = ({
  title,
  description,
  extra,
  children,
}) => {
  return (
    <div>
      {/* 页面头部 */}
      <div
        className="page-header"
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'space-between',
        }}
      >
        <div>
          <div className="page-title">{title}</div>
          {description && <div className="page-desc">{description}</div>}
        </div>
        {extra && <div style={{ flexShrink: 0, marginLeft: 24 }}>{extra}</div>}
      </div>

      {/* 页面内容 */}
      {children}
    </div>
  );
};

export default PageContainer;
