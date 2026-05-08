# 工作流废弃内容检查报告

**检查日期**: 2026-05-08
**检查范围**: 所有 GitHub Actions 工作流
**检查内容**: 废弃字段、废弃机制、过时信息、未使用内容

---

## 检查结果总结

| 工作流 | 废弃字段 | 废弃机制 | 过时信息 | 未使用内容 | 状态 |
|--------|----------|----------|----------|------------|------|
| cd-workflow.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ⚠️ sign-artifacts | ⚠️ 需要修复 |
| ci.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ⚠️ docs job | ⚠️ 需要修复 |
| dev-workflow.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ❌ 无 | ✅ 正常 |
| document-audit.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ❌ 无 | ✅ 正常 |
| monitoring-workflow.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ❌ 无 | ✅ 正常 |
| pr-check-workflow.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ❌ 无 | ✅ 正常 |
| release.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ⚠️ 与 cd-workflow 重复 | ⚠️ 需要修复 |
| security-workflow.yml | ❌ 无 | ❌ 无 | ✅ 正确 | ❌ 无 | ✅ 正常 |

---

## 详细问题分析

### 1. 过时的 Go 版本

**影响范围**: 所有工作流
**当前版本**: 1.26
**状态**: ✅ 正确版本
**说明**: 项目使用 Go 1.26 版本（用户确认），该版本配置正确

---

### 2. 未使用的内容

#### 2.1 cd-workflow.yml - sign-artifacts job

**位置**: cd-workflow.yml lines 114-141
**问题**: sign-artifacts job 生成签名但没有被 create-release job 使用
**详情**:
```yaml
sign-artifacts:
  name: Sign Artifacts
  runs-on: ubuntu-latest
  needs: [build-multi-platform, build-docker]
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    - name: Download all artifacts
      uses: actions/download-artifact@v4
      with:
        path: ./artifacts
    - name: Sign binaries
      env:
        GPG_PRIVATE_KEY: ${{ secrets.GPG_PRIVATE_KEY }}
        GPG_PASSPHRASE: ${{ secrets.GPG_PASSPHRASE }}
      run: |
        echo "$GPG_PRIVATE_KEY" | gpg --import --batch --passphrase "$GPG_PASSPHRASE"
        for file in artifacts/axis-*; do
          gpg --detach-sign --batch --passphrase "$GPG_PASSPHRASE" "$file"
        done
    - name: Upload signatures
      uses: actions/upload-artifact@v4
      with:
        name: signatures
        path: artifacts/*.asc
```

**修复建议**:
1. 在 create-release job 中下载并上传签名文件
2. 或者删除 sign-artifacts job（如果不需要签名）

#### 2.2 ci.yml - docs job

**位置**: ci.yml lines 143-167
**问题**: docs job 生成 API 文档但没有在其他地方使用
**详情**:
```yaml
docs:
  name: Generate Documentation
  runs-on: ubuntu-latest
  needs: test
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.26'
    - name: Generate API documentation
      run: |
        mkdir -p docs/api
        go doc -all . > docs/api/index.txt
        echo "Documentation generated at docs/api/index.txt"
    - name: Upload documentation artifact
      uses: actions/upload-artifact@v4
      with:
        name: documentation
        path: docs/api/
        retention-days: 7
```

**修复建议**:
1. 将文档上传到 GitHub Pages
2. 或者删除 docs job（如果不需要自动生成文档）

#### 2.3 release.yml 与 cd-workflow.yml 重复

**问题**: release.yml 和 cd-workflow.yml 都执行发布功能，存在重复
**详情**:
- release.yml: 简单的发布流程
- cd-workflow.yml: 更完整的发布流程（包括 Docker 构建和多平台二进制）

**修复建议**:
1. 删除 release.yml，统一使用 cd-workflow.yml
2. 或者将 release.yml 重命名为 legacy-release.yml 并标记为废弃

---

### 3. 废弃字段检查

**结果**: 未发现废弃字段

所有使用的 GitHub Actions 版本都是当前版本：
- actions/checkout@v4 ✅
- actions/setup-go@v5 ✅
- docker/setup-buildx-action@v3 ✅
- docker/login-action@v3 ✅
- docker/build-push-action@v5 ✅
- actions/upload-artifact@v4 ✅
- actions/download-artifact@v4 ✅
- softprops/action-gh-release@v1 ✅
- codecov/codecov-action@v4 ✅
- benchmark-action/github-action-benchmark@v1 ✅
- actions/github-script@v7 ✅
- trufflesecurity/trufflehog@main ✅

---

### 4. 废弃机制检查

**结果**: 未发现废弃机制

所有使用的机制都是当前支持的机制：
- workflow_dispatch ✅
- schedule ✅
- workflow_run ✅
- matrix strategy ✅
- needs dependency ✅
- permissions ✅
- artifacts ✅
- secrets ✅

---

### 5. 过时信息检查

**结果**: 发现过时的 Go 版本信息

详见第 1 节"过时的 Go 版本"

---

## 修复优先级

### 高优先级（必须修复）
1. **修复*修复 Go** - 所有工作流
   - 原因: 使用不存在的版本会导致工作流失败
   - 影响: 所有工作流都无法*常运行- 所有工作流
   - 原因: 使用不存在的版本会导致工作流失败
   - 影响: 所有工作流都无法正常运行
2
### 中优先级（建议修复）
2. **修复 cd-workflow.yml 的 sign-artifacts job**
   - 原因: 生成签名但不使用，浪费资源
3  - 影响: CI/CD 时间增加

3. **修复 ci.yml 的 docs job**
   - 原因: 生成文档但不使用，浪费资源
4  - 影响: CI/CD 时间增加

4. **解决 release.yml 和 cd-workflow.yml 重复**
   - 原因: 维护两套发布流程容易出错
   - 影响: 维护成本增加

---

## 修复建议

### 建议 1: 修复 cd-workflow.yml 的 sign-artifacts

**选项 A**: 在 create-release 中使用签名
```yaml
- name: Download signatures
  uses: actions/download-artifact@v4
  with:
    name: signatures
    path: ./signatures

- name: Create Release
  uses: softprops/action-gh-release@v1
  with:
    files: |
      artifacts/*
      artifacts/SHA256SUMS.txt
      signatures/*.asc
```

**选项 B**: 删除 sign-artifacts job
如果不需要签名，直接删除该 job

### 建议 3: 修复 ci.yml 的 docs job

**选项 A**: 上传到 GitHub Pages
```yaml
- name: Deploy to GitHub Pages
  uses: peaceiris/actions-gh-pages@v3
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    publish_dir: ./docs/api
```2

**选项 B**: 删除 docs job
如果不需要自动生成文档，直接删除该 job

### 建议 4: 统一发布流程

删除 release.yml，统一使用 cd-workflow.yml 进行发布

---

## 执行计划
3
### 阶段 1: 修复 Go 版本（立即执行）
1. 更新所有工作流的 Go 版本为 1.22
2. 测试 CI workflow
3. 提交并推送

### 阶段 2: 修复未使用内容（本周）
1. 决定 sign-artifacts 的处理方式
2. 决定 docs job 的处理方式
3. 决定发布流程的统一方式
4. 实施修复
5. 测试
6. 提交并推送

---

## 结论

**总体状态**: ⚠️ 需要优化

**主要问题**:
1. cd-workflow.yml 的 sign-artifacts job 未被使用
2. ci.yml 的 docs job 未被使用
3. release.yml 和 cd-workflow.yml 功能重复

**建议行动**:
1. 本周内修复未使用内容问题
2. 统一发布流程

**预期结果**:
修复后，资源使用更加高效，维护成本降低。
