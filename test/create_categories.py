#!/usr/bin/env python3

import requests
import json
from typing import List, Dict, Optional

BASE_URL = "http://localhost:8888/api"


class CategoryCreator:
    def __init__(self, base_url: str = BASE_URL):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})

    def create_category(
        self,
        name: str,
        description: str = "",
        parent_id: Optional[int] = None,
        position: int = 0,
    ) -> Dict:
        payload = {"name": name, "description": description, "position": position}
        if parent_id is not None:
            payload["parent_id"] = parent_id

        try:
            response = self.session.post(f"{self.base_url}/categories", json=payload)
            response.raise_for_status()
            result = response.json()
            print(f"✓ 创建分类成功: {name} (ID: {result['id']})")
            return result
        except requests.exceptions.RequestException as e:
            print(f"✗ 创建分类失败: {name} - {e}")
            if hasattr(e.response, "text"):
                print(f"  错误详情: {e.response.text}")
            return None

    def get_categories(self) -> List[Dict]:
        try:
            response = self.session.get(f"{self.base_url}/categories")
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"✗ 获取分类列表失败: {e}")
            return []


def main():
    creator = CategoryCreator()

    # 首先创建根分类
    root_categories = [
        {"name": "时间线", "description": "时间线相关内容", "position": 1},
        {"name": "综合", "description": "综合类内容", "position": 2},
        {"name": "水下芦苇", "description": "水下芦苇相关内容", "position": 3},
        {"name": "管理", "description": "管理相关内容", "position": 4},
        {"name": "博物馆", "description": "博物馆相关内容", "position": 5},
    ]

    # 子分类（需要在根分类创建后关联）
    child_categories = [
        {
            "name": "水上芦苇",
            "description": "水上芦苇相关内容",
            "position": 1,
            "parent": "时间线",
        },
        {
            "name": "欢乐恶搞",
            "description": "欢乐恶搞内容",
            "position": 1,
            "parent": "综合",
        },
        {
            "name": "围炉",
            "description": "围炉相关内容",
            "position": 2,
            "parent": "综合",
        },
        {"name": "创意", "description": "创意内容", "position": 3, "parent": "综合"},
        {
            "name": "跑团",
            "description": "跑团相关内容",
            "position": 4,
            "parent": "综合",
        },
        {
            "name": "故事(小说)",
            "description": "故事和小说",
            "position": 5,
            "parent": "综合",
        },
        {
            "name": "询问",
            "description": "询问相关内容",
            "position": 6,
            "parent": "综合",
        },
        {"name": "怪谈", "description": "怪谈内容", "position": 7, "parent": "综合"},
        {
            "name": "游戏",
            "description": "游戏相关内容",
            "position": 8,
            "parent": "综合",
        },
        {
            "name": "动画",
            "description": "动画相关内容",
            "position": 9,
            "parent": "综合",
        },
        {
            "name": "漫画",
            "description": "漫画相关内容",
            "position": 10,
            "parent": "综合",
        },
        {
            "name": "购物",
            "description": "购物相关内容",
            "position": 11,
            "parent": "综合",
        },
        {
            "name": "科技",
            "description": "科技相关内容",
            "position": 12,
            "parent": "综合",
        },
        {
            "name": "料理(宠物)",
            "description": "料理和宠物相关内容",
            "position": 13,
            "parent": "综合",
        },
        {
            "name": "东方Project",
            "description": "东方Project相关内容",
            "position": 14,
            "parent": "综合",
        },
        {
            "name": "Minecraft",
            "description": "Minecraft相关内容",
            "position": 15,
            "parent": "综合",
        },
        {
            "name": "社畜",
            "description": "社畜相关内容",
            "position": 16,
            "parent": "综合",
        },
        {
            "name": "学业",
            "description": "学业相关内容",
            "position": 17,
            "parent": "综合",
        },
    ]

    category_map = {}

    print("=" * 50)
    print("开始创建分类数据...")
    print("=" * 50)

    # 创建根分类
    print("创建根分类...")
    for cat in root_categories:
        result = creator.create_category(
            name=cat["name"], description=cat["description"], position=cat["position"]
        )
        if result:
            category_map[cat["name"]] = result["id"]

    # 创建子分类
    print("\n创建子分类...")
    for cat in child_categories:
        parent_id = category_map.get(cat["parent"])
        if parent_id is None:
            print(f"✗ 找不到父分类: {cat['parent']}")
            continue

        result = creator.create_category(
            name=cat["name"],
            description=cat["description"],
            position=cat["position"],
            parent_id=parent_id,
        )
        if result:
            category_map[cat["name"]] = result["id"]

    print("=" * 50)
    print("创建完成！")
    print("=" * 50)

    print("\n创建的分类映射:")
    for name, cat_id in category_map.items():
        print(f"  {name}: {cat_id}")

    print("\n获取所有分类树:")
    all_categories = creator.get_categories()
    print(json.dumps(all_categories, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
