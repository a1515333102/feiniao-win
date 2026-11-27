<template>
  <div class="最底层div">
    <el-form v-loading="is加载中" :inline="false" style="min-width: 80px;display: flex;flex-direction: column" label-width="130px" :model="Data"
             :label-position="is移动端()?'top':'right'" ref="ruleFormRef">
      <div class="内容div" :style="{order: Data.余额支付排序}">
        <el-divider content-position="left">
          <el-button-group class="排序按钮组">
            <el-button size="small" :disabled="Q是否已到顶部('余额支付排序')" title="上移" @click="on移动排序('余额支付排序', -1)">
              <el-icon>
                <ArrowUp/>
              </el-icon>
            </el-button>
            <el-button size="small" :disabled="Q是否已到底部('余额支付排序')" title="下移" @click="on移动排序('余额支付排序', 1)">
              <el-icon>
                <ArrowDown/>
              </el-icon>
            </el-button>
          </el-button-group>
          余额支付
        </el-divider>
        <el-form-item label="开关" prop="余额支付开关">
          <el-radio-group v-model="Data.余额支付开关">
            <el-radio-button :value="true" size="" border>开启</el-radio-button>
            <el-radio-button :value="false" size="" border>关闭</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <div v-if="Data.余额支付开关">
          <el-form-item label="显示名称" disabled="disabled">
            <el-input v-model.trim="Data.余额支付显示名称"/>
          </el-form-item>
        </div>
      </div>

      <div class="内容div" :style="{order: Data.易支付排序}">
        <el-divider content-position="left">
          <el-button-group class="排序按钮组">
            <el-button size="small" :disabled="Q是否已到顶部('易支付排序')" title="上移" @click="on移动排序('易支付排序', -1)">
              <el-icon>
                <ArrowUp/>
              </el-icon>
            </el-button>
            <el-button size="small" :disabled="Q是否已到底部('易支付排序')" title="下移" @click="on移动排序('易支付排序', 1)">
              <el-icon>
                <ArrowDown/>
              </el-icon>
            </el-button>
          </el-button-group>
          在线支付（易支付协议）
          <el-text>-------</el-text>
          <el-link href="https://pay.10sn.cn/user" target="_blank">商户中心</el-link>
        </el-divider>
        <el-form-item label="开关" prop="易支付开关">
          <el-radio-group v-model="Data.易支付开关">
            <el-radio-button :value="true" size="" border>开启</el-radio-button>
            <el-radio-button :value="false" size="" border>关闭</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <div v-if="Data.易支付开关">
          <el-form-item label="显示名称" disabled="disabled">
            <el-input v-model.trim="Data.易支付显示名称"/>
          </el-form-item>
          <el-form-item label="网关" disabled="disabled" placeholder="https://www.baidu.com">
            <el-input v-model.trim="Data.易支付网关"/>
          </el-form-item>
          <el-form-item label="商户ID" disabled="disabled">
            <el-input v-model.trim="Data.易支付商户ID"/>
          </el-form-item>
          <el-form-item label="接口密钥" disabled="disabled">
            <el-input v-model.trim="Data.易支付商户密钥KEY"/>
          </el-form-item>
          <el-form-item label="支付方式" disabled="disabled">
            <el-input v-model.trim="Data.易支付支付方式"/>
          </el-form-item>
          <el-form-item label="设备类型" disabled="disabled">
            <el-input v-model.trim="Data.易支付设备类型" placeholder="(可空)自动根据请求ua识别"/>
          </el-form-item>
          <el-form-item label="同步回调url" d>
            <el-input v-model.trim="Data.易支付同步回调url" placeholder="(可空)同步回调地址支持变量`{OrderId}`订单id,`{OrderId2}`第三方订单id,`{User}`用户名,`{Type}`支付方式"/>
          </el-form-item>
          <el-form-item label="单次最大金额" disabled="disabled">
            <el-input-number v-model="Data.易支付最大金额"/>
          </el-form-item>
          <div style="padding: 0 0 8px 8px;font-size: 12px;color: #909399">
            异步回调可不配：系统每 30 秒主动向上游查询一次订单状态，确认已支付后自动入账。
          </div>
        </div>
      </div>



      <div class="内容div" style="order: 9999">
        <div style="text-align:center">
          <el-button style="width: 15vh; " type="primary" @click="on确定按钮被点击(ruleFormRef)">保存</el-button>
        </div>
      </div>
    </el-form>
  </div>
</template>

<script lang="ts" setup>
import {onBeforeUnmount, onMounted, ref} from "vue";
import {GetInfoPay, SaveInfoPay} from "@/api/系统设置api.js";
import {ElMessage, FormInstance} from 'element-plus'
import {is移动端} from "@/utils/utils";

const Data = ref({
  "易支付开关": false,
  "易支付显示名称": "",
  "易支付网关": "https://pay.10sn.cn",
  "易支付商户ID": "",
  "易支付商户密钥KEY": "",
  "易支付支付方式": "alipay",
  "易支付设备类型": "",
  "易支付同步回调url": "",
  "易支付最大金额": 2000,

  "余额支付开关": false,
  "余额支付显示名称": "",

  "易支付排序": 1,
  "余额支付排序": 2,
})

const ruleFormRef = ref<FormInstance>()
const is加载中 = ref(false)

//支付方式排序字段列表,顺序即缺省排序顺序
const 排序字段列表 = [
  "易支付排序",
  "余额支付排序",
] as const

//排序缺省值补全,旧配置可能没有排序值
const on排序缺省值补全 = () => {
  排序字段列表.forEach((字段, 索引) => {
    if (!Data.value[字段]) {
      Data.value[字段] = 索引 + 1
    }
  })
}

//取按排序值升序排列后的字段列表
const Q取排序后字段列表 = () => {
  return 排序字段列表
      .map(字段 => ({字段: 字段, 排序: Number(Data.value[字段]) || 0}))
      .sort((a, b) => a.排序 - b.排序)
}

//上移一位(-1)或下移一位(1),与相邻支付方式交换排序值
const on移动排序 = (字段名: typeof 排序字段列表[number], 方向: number) => {
  const 列表 = Q取排序后字段列表()
  const 局_当前索引 = 列表.findIndex(v => v.字段 === 字段名)
  const 局_目标索引 = 局_当前索引 + 方向
  if (局_当前索引 < 0 || 局_目标索引 < 0 || 局_目标索引 >= 列表.length) {
    return
  }
  const 局_临时 = Data.value[列表[局_当前索引].字段]
  Data.value[列表[局_当前索引].字段] = Data.value[列表[局_目标索引].字段]
  Data.value[列表[局_目标索引].字段] = 局_临时
}

const Q是否已到顶部 = (字段名: string) => Q取排序后字段列表()[0].字段 === 字段名
const Q是否已到底部 = (字段名: string) => {
  const 列表 = Q取排序后字段列表()
  return 列表[列表.length - 1].字段 === 字段名
}

const onGetList = async () => {
  is加载中.value = true
  const res = await GetInfoPay({})
  Data.value = res.data
  on排序缺省值补全()
  is加载中.value = false
}

onMounted(async () => {
  await onGetList()
})
const on确定按钮被点击 = async (formEl: FormInstance | undefined) => {


  if (!formEl) return
  let 表单验证结果 = await formEl.validate((valid, fields) => {
    if (!valid) {
      console.log('参数验证失败', fields)
    } else {
      console.log('参数验证通过')
    }
  })
  console.info("表单验证结果")
  console.info(表单验证结果)
  if (!表单验证结果) return   //如果是假直接返回
  let 返回;
  is加载中.value = true
  返回 = await SaveInfoPay(Data.value);
  is加载中.value = false
  console.log(返回)
  if (返回.code == 10000) {
ElMessage.success(返回.msg)
  }
}
</script>

<style scoped lang="scss">
.el-table .cell {
  white-space: pre-wrap; /*这是重点。文本换行*/

}

/*.gva-search-box {*/
/*  padding: 24px;*/
/*  padding-bottom: 2px;*/
/*  background-color: #fff;*/
/*  border-radius: 2px;*/
/*  margin-bottom: 12px;*/
/*}*/
.最底层div {

  padding: 12px 16px;
  margin: 0 2px 10px;
  background: #f0f2f5;
}

.内容div {
  min-height: 20%;
  padding: 12px 16px;
  margin: 0 2px 10px;
  background: #ffffff;
}

.排序按钮组 {
  margin-right: 8px;
}

.搜索框 {
  top: -5px;
  padding: 0 0;
  margin: 0 0 10px;
  align-items: center;
}

.demo-pagination-block {
  margin-top: 15px;
  display: flex;
  justify-content: flex-end;
}

.el-statistic__number {
  font-size: 18px;
  color: #eebe77;
}

.gva-btn-list {
  border: 1px solid rgb(235, 238, 245);
}

.工具栏 {
  margin: 7px 8px 8px;
  background: #fafafa;
  color: #606266;
  float: right;
  padding-right: 1px;

  .el-icon {
    /*设置边框阴影*/

    font-size: 16px;
    margin-left: 10px;
    padding: 5px;
    ///*边框 1px  颜色 */
    border: 1px solid rgb(235, 238, 245);
    color: #0c0d0e;
    //box-shadow: 2px 2px 3px 0 rgba(45, 75, 74, 0.6);
    speak: none;
    font-style: normal;
    font-variant: normal;
    text-transform: none;
    line-height: 1;
    vertical-align: baseline;
    display: inline-block;
    -webkit-font-smoothing: antialiased;
    cursor: pointer; //改变鼠标样式为手型
  }
}


.工具_更多 {
  background-color: #ffffff;
  width: 150px;
  margin: 0;
  /*边框 1px  颜色 */
  border: 1px solid #ccc;
  /*图层高度  3000  值大一点 会在顶层*/
  z-index: 3000;
  /*定位方式 绝对定位*/
  position: absolute;
  list-style-type: none;
  border-radius: 4px; //设置圆角
  /*设置边框阴影*/
  box-shadow: 2px 2px 3px 0 rgba(45, 75, 74, 0.6);
  padding: 5px 0;
  font-size: 12px;

  li {
    margin: 0;
    padding: 7px 16px;
    //设置 鼠标悬停时样式
    &:hover {
      background: #889aa4; //改变背景颜色
      cursor: pointer; //改变鼠标样式为手型
    }
  }

}

.工具_更多_li {
  list-style-type: none;
  font-size: 12px;
  margin: 0;
  padding: 7px 16px;
  //设置 鼠标悬停时样式
  &:hover {
    background: #889aa4; //改变背景颜色
    cursor: pointer; //改变鼠标样式为手型
  }
}

.el-form-item {
  padding: 0;
  margin: 0 15px 8px 0;
}

.el-table .cell {
  white-space: pre-line;
}

.编辑框 {
  width: 200px;
}

单条信息分组框 {
  padding: 20px 10px 10px 10px;

}
</style>
