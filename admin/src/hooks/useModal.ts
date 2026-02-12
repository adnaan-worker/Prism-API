import { useState } from 'react';
import { Form } from 'antd';

/**
 * 模态框状态管理 Hook
 * 统一管理模态框的显示/隐藏、表单等状态
 */
export function useModal<T = any>() {
  const [visible, setVisible] = useState(false);
  const [editingItem, setEditingItem] = useState<T | null>(null);
  const [form] = Form.useForm();

  const showModal = (item?: T) => {
    if (item) {
      setEditingItem(item);
      form.setFieldsValue(item);
    } else {
      setEditingItem(null);
      form.resetFields();
    }
    setVisible(true);
  };

  const hideModal = () => {
    setVisible(false);
    setEditingItem(null);
    form.resetFields();
  };

  const isEditing = editingItem !== null;

  return {
    visible,
    editingItem,
    form,
    isEditing,
    showModal,
    hideModal,
    // 兼容旧的命名
    isModalVisible: visible,
  };
}
