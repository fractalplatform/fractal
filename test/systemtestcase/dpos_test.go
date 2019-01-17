// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRegProducer(t *testing.T) {
	Convey("注册生成者", t, func() {
		Convey("生产者", func() {
			Convey("已是投票者", func() {

			})
			Convey("已是生产者", func() {

			})
		})
		Convey("URL长度", func() {
			Convey("不满足", func() {

			})
			Convey("满足", func() {

			})
		})
		Convey("最低数量要求", func() {
			Convey("不满足", func() {

			})
			Convey("满足", func() {

			})
		})
	})
}

func TestUpdateProducer(t *testing.T) {
	Convey("更新生成者", t, func() {
		Convey("生产者", func() {
			Convey("已是投票者", func() {

			})
			Convey("不是生产者", func() {

			})
		})
		Convey("URL长度", func() {
			Convey("不满足", func() {

			})
			Convey("满足", func() {

			})
		})
		Convey("最低数量要求", func() {
			Convey("不满足", func() {

			})
			Convey("满足", func() {

			})
		})
	})
}

func TestUnRegProducer(t *testing.T) {
	Convey("注销生成者", t, func() {
		Convey("生产者", func() {
			Convey("已是投票者", func() {

			})
			Convey("不是生产者", func() {

			})
		})
		Convey("URL长度", func() {
			Convey("不满足", func() {

			})
			Convey("满足", func() {

			})
		})
		Convey("最低数量要求", func() {
			Convey("不满足", func() {

			})
			Convey("满足", func() {

			})
		})
	})
}
