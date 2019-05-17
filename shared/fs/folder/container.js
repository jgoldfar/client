// @flow
import * as I from 'immutable'
import {namedConnect} from '../../util/container'
import Folder from '.'
import * as Types from '../../constants/types/fs'
import * as Constants from '../../constants/fs'
import * as FsGen from '../../actions/fs-gen'

const mapStateToProps = (state, {path}) => ({
  _kbfsDaemonStatus: state.fs.kbfsDaemonStatus,
  _pathItem: state.fs.pathItems.get(path, Constants.unknownPathItem),
  resetBannerType: Constants.resetBannerType(state, path),
  shouldShowSFMIBanner: state.fs.sfmi.showingBanner,
  username: state.config.username,
})

const mapDispatchToProps = (dispatch, {path}: OwnProps) => ({
  onAttach: (paths: Array<string>) => {
    paths.forEach(localPath => dispatch(FsGen.createUpload({localPath, parentPath: path})))
  },
})

const mergeProps = (stateProps, dispatchProps, {path, routePath}) => ({
  offline: Constants.isOfflineUnsynced(stateProps._kbfsDaemonStatus, stateProps._pathItem, path),
  onAttach: stateProps._pathItem.writable ? dispatchProps.onAttach : null,
  path,
  resetBannerType: stateProps.resetBannerType,
  routePath,
  shouldShowSFMIBanner: stateProps.shouldShowSFMIBanner,
  username: stateProps.username,
})

type OwnProps = {|
  path: Types.Path,
  routePath: I.List<string>,
|}

// flow can't figure out type when compose is used.
export default namedConnect<OwnProps, _, _, _, _>(mapStateToProps, mapDispatchToProps, mergeProps, 'Folder')(
  Folder
)
